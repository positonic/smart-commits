# Smart Git Commit

A command-line tool that uses OpenAI's GPT-3.5 to automatically generate meaningful git commit messages based on your changes.

## Features

- Automatically generates conventional commit messages from your git diff
- Works with both staged and unstaged changes
- Interactive confirmation before committing
- Automatically stages all changes upon confirmation

## Installation

1. Clone this repository
2. Build this binary: `go build -o smart-commit`
3. Move the binary to your path: `mv smart-commit ~/bin/`
4. Set your OpenAI API key as an environment variable:
```bash
export OPENAI_API_KEY='your-api-key-here'
```

To persist your API key, you can add it to your `.zshrc` file:
```bash
echo "export OPENAI_API_KEY='your-api-key-here'" >> ~/.zshrc
```

## Usage
Navigate to your git project:
```bash
cd your-project
```

Make some changes to your files

Run the tool:

```bash
smart-commit  # Go version
```

I use an alias in my zsh config to make it easier to run:

```bash
echo 'alias sc="smart-commit"' >> ~/.zshrc
source ~/.zshrc
```

and run it with `sc`


The script will:
1. Get the diff of your changes (staged or unstaged)
2. Send it to OpenAI's API
3. Generate a commit message
4. Show you the proposed message
5. Ask for confirmation before proceeding
6. If confirmed, stage all changes and create the commit

## Current Limitations

- Requires all changes to be committed together (no partial commits)
- Requires manual API key setup

## Contributing

Contributions are welcome! Some potential areas for improvement:
- Add support for detailed commit messages with body and footer
- Add configuration options for commit message style
- Add support for partial commits
- Add rate limiting and token usage optimization
- Add support for different AI models or providers

## License

This project is open-sourced under the MIT License - see the LICENSE file for details.
