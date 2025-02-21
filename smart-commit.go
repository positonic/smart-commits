package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func main() {
	// Check for OpenAI API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY is not set")
		os.Exit(1)
	}

	// Get git diff
	diff := getGitDiff()
	if diff == "" {
		fmt.Println("No changes to commit")
		os.Exit(1)
	}

	fmt.Println("Generating commit message...")

	// Create OpenAI request
	req := OpenAIRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{
				Role: "system",
				Content: "You are a git commit message writer. Generate a SINGLE commit message that summarizes all changes. Format the message as follows:\n\n" +
					"<type>: <subject>\n\n" +
					"- <change 1>\n" +
					"- <change 2>\n" +
					"- <change 3>\n\n" +
					"Rules:\n" +
					"- Create ONE commit message that encompasses all changes\n" +
					"- First line must be under 50 chars and follow conventional commit format\n" +
					"- Each bullet point should start with a capital letter and be a single line\n" +
					"- Bullet points should be clear and concise\n" +
					"- Use types: feat, fix, docs, style, refactor, test, chore\n" +
					"- Focus on WHAT changed and WHY, not HOW\n" +
					"- If multiple types of changes exist, choose the most significant type",
			},
			{
				Role:    "user",
				Content: diff,
			},
		},
	}

	// Get commit message from OpenAI
	commitMsg := getCommitMessage(req, apiKey)

	// Show the message and get confirmation
	fmt.Printf("\nProposed commit message:\n%s\n", commitMsg)
	fmt.Print("\nDo you want to proceed with this commit message? (y/n): ")

	var confirm string
	fmt.Scanln(&confirm)

	if strings.ToLower(confirm) == "y" {
		// Stage and commit changes
		execCommand("git", "add", ".")
		execCommand("git", "commit", "-m", commitMsg)
		fmt.Println("\nSuccessfully committed!")
	} else {
		fmt.Println("\nCommit cancelled")
	}
}

func getGitDiff() string {
	diff := execCommand("git", "diff")
	if diff == "" {
		diff = execCommand("git", "diff", "--cached")
	}
	return diff
}

func getCommitMessage(req OpenAIRequest, apiKey string) string {
	// Convert request to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}

	client := &http.Client{}
	httpReq, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Printf("Error sending request to OpenAI: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	output := body
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized:
		fmt.Printf("Error: Invalid API key or unauthorized access (Status: %d)\n", resp.StatusCode)
		fmt.Printf("Response: %s\n", string(body))
		os.Exit(1)
	case http.StatusBadRequest:
		fmt.Printf("Error: Bad request - check your input parameters (Status: %d)\n", resp.StatusCode)
		fmt.Printf("Response: %s\n", string(body))
		os.Exit(1)
	case http.StatusTooManyRequests:
		fmt.Printf("Error: Rate limit exceeded (Status: %d)\n", resp.StatusCode)
		fmt.Printf("Response: %s\n", string(body))
		os.Exit(1)
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		fmt.Printf("Error: OpenAI service error (Status: %d)\n", resp.StatusCode)
		fmt.Printf("Response: %s\n", string(body))
		os.Exit(1)
	default:
		fmt.Printf("Error: Unexpected status code: %d\n", resp.StatusCode)
		fmt.Printf("Response: %s\n", string(body))
		os.Exit(1)
	}

	var response OpenAIResponse
	if err := json.Unmarshal(output, &response); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		fmt.Printf("Raw response: %s\n", string(output))
		os.Exit(1)
	}

	if len(response.Choices) == 0 {
		fmt.Println("No commit message generated")
		os.Exit(1)
	}

	return response.Choices[0].Message.Content
}

func execCommand(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		os.Exit(1)
	}
	return out.String()
}
