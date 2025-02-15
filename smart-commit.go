package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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
                Role:    "system",
                Content: "You are a git commit message writer. Format commit messages as follows:\n\n" +
                    "<type>: <subject>\n" +
                    "- <change 1>\n" +
                    "- <change 2>\n" +
                    "- <change 3>\n\n" +
                    "Rules:\n" +
                    "- First line must be under 50 chars and follow conventional commit format\n" +
                    "- Each bullet point should start with a capital letter and be a single line\n" +
                    "- Bullet points should be clear and concise\n" +
                    "- Use types: feat, fix, docs, style, refactor, test, chore\n" +
                    "- Focus on WHAT changed and WHY, not HOW",
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

    // Create curl command
    cmd := exec.Command("curl", 
        "-s",
        "https://api.openai.com/v1/chat/completions",
        "-H", "Content-Type: application/json",
        "-H", fmt.Sprintf("Authorization: Bearer %s", apiKey),
        "-d", string(jsonData))

    output, err := cmd.CombinedOutput()
    if err != nil {
        fmt.Printf("Error calling OpenAI API: %v\n", err)
        os.Exit(1)
    }

    // Parse response
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