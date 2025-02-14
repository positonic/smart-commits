# Smart Git Commit

A command-line tool that uses OpenAI's GPT-3.5 to automatically generate meaningful git commit messages based on your changes. [Learn more](https://www.jamesfarrell.me/blog/autogen-git-commit-messages#go-implementation)

## Features

- Automatically generates conventional commit messages from your git diff
- Works with both staged and unstaged changes
- Interactive confirmation before committing
- Automatically stages all changes upon confirmation

## Prerequisites

- Zsh shell
- Git
- `jq` command-line JSON processor
- OpenAI API key

## Installation

1. Clone this repository
2. Set your OpenAI API key as an environment variable:
```bash
export OPENAI_API_KEY='your-api-key-here'
```

**For the Go version (recommended)**
Build the binary:

```bash
go build -o smart-commit
```

Move to your bin directory:

```bash
mv smart-commit ~/bin/
```

Set your OpenAI API key (add to your shell config for persistence):
echo 'export OPENAI_API_KEY="your-key-here"' >> ~/.zshrc

```bash
source ~/.zshrc
```

## Usage of bash version:
```bash
bash
./bin/smart-commit.zsh
```


The script will:
1. Get the diff of your changes (staged or unstaged)
2. Send it to OpenAI's API
3. Generate a commit message
4. Show you the proposed message
5. Ask for confirmation before proceeding
6. If confirmed, stage all changes and create the commit

## Current Limitations

- Currently only generates the commit subject line (no detailed body or footer)
- Uses GPT-3.5-turbo model (may incur API costs)
- Requires all changes to be committed together (no partial commits)
- No configuration options for commit message style or conventions
- Requires manual API key setup
- No handling of large diffs that might exceed API token limits

## Contributing

Contributions are welcome! Some potential areas for improvement:
- Add support for detailed commit messages with body and footer
- Add configuration options for commit message style
- Add support for partial commits
- Add rate limiting and token usage optimization
- Add support for different AI models or providers

## License

This project is open-sourced under the MIT License - see the LICENSE file for details.
