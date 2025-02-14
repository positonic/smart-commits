#!/bin/zsh

# Check if OPENAI_API_KEY is set
if [ -z "$OPENAI_API_KEY" ]; then
    echo "Error: OPENAI_API_KEY is not set"
    exit 1
fi

# Get the git diff
DIFF=$(git diff)

# If no diff, check for staged changes
if [ -z "$DIFF" ]; then
    DIFF=$(git diff --cached)
    if [ -z "$DIFF" ]; then
        echo "No changes to commit"
        exit 1
    fi
fi

# Escape the diff content for JSON
DIFF_ESCAPED=$(echo "$DIFF" | jq -sR .)

echo "Generating commit message..."

# Use curl to send to OpenAI API with escaped diff
RESPONSE=$(curl -s https://api.openai.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {
        "role": "system",
        "content": "You are a helpful git commit message writer. Given a git diff, write a concise, conventional commit message. Return ONLY the commit message, no explanation or markdown."
      },
      {
        "role": "user",
        "content": '"$DIFF_ESCAPED"'
      }
    ]
  }')

# Add debug lines here
echo "Debug - Raw API Response:"
echo "$RESPONSE"

# Check if curl request was successful
if [ $? -ne 0 ]; then
    echo "Error: Failed to connect to OpenAI API"
    exit 1
fi

# Extract the commit message from the JSON response
COMMIT_MSG=$(echo $RESPONSE | jq -r '.choices[0].message.content')

# Check if jq parsing was successful
if [ $? -ne 0 ]; then
    echo "Error: Failed to parse API response"
    echo "Raw response: $RESPONSE"
    exit 1
fi

# Show the commit message and ask for confirmation
echo "\nProposed commit message:\n$COMMIT_MSG"
echo "\nDo you want to proceed with this commit message? (y/n)"
read CONFIRM

if [[ $CONFIRM =~ ^[Yy]$ ]]; then
    # Stage all changes
    git add .
    
    # Commit with the generated message
    git commit -m "$COMMIT_MSG"
    
    echo "\nSuccessfully committed with message: $COMMIT_MSG"
else
    echo "\nCommit cancelled"
    exit 0
fi