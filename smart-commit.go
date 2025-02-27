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

// Maximum size for diff in bytes (approximately 12K tokens)
const maxDiffSize = 48000

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

    if len(diff) > maxDiffSize {
        // Get summary information
        filesSummary := execCommand("git", "diff", "--stat")
        if filesSummary == "" {
            filesSummary = execCommand("git", "diff", "--cached", "--stat")
        }

        // Get truncated diff with --unified=1 for more concise output
        truncatedDiff := execCommand("git", "diff", "--unified=1")
        if truncatedDiff == "" {
            truncatedDiff = execCommand("git", "diff", "--cached", "--unified=1")
        }

        // If still too large, take the first portion
        if len(truncatedDiff) > maxDiffSize {
            truncatedDiff = truncatedDiff[:maxDiffSize]
        }

        return fmt.Sprintf("Summary of changes:\n%s\n\nPartial diff (truncated due to size):\n%s",
            filesSummary, truncatedDiff)
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