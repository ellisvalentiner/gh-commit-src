# gh-commit-src

> Automated Git commit messages powered by LLMs

[![CI](https://github.com/ellisvalentiner/gh-commit-src/actions/workflows/push.yml/badge.svg)](https://github.com/ellisvalentiner/gh-commit-src/actions/workflows/push.yml)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**gh-commit-src** is a GitHub CLI extension that automatically generates meaningful commit messages using Large Language Models (LLMs). It analyzes your code changes and creates commit messages following the [Conventional Commits](https://www.conventionalcommits.org/) specification.

> **Note:** This is a fork of [github.com/megamanics/gh-commit](https://github.com/megamanics/gh-commit) with improvements and modernizations.

## Features

- 🤖 **AI-Powered**: Uses OpenAI or Azure OpenAI to generate contextual commit messages
- 📝 **Conventional Commits**: Automatically formats messages following industry standards
- 🔧 **Configurable**: Customize prompts, model parameters, and API endpoints
- 🚀 **Easy Integration**: Works as both a Git alias and GitHub CLI extension
- ⚡ **Fast**: Generates commit messages in seconds

## Prerequisites

- **Go** 1.21 or later
- **Git** 2.30 or later
- **GitHub CLI** 2.0 or later
- **OpenAI API Key** (or Azure OpenAI endpoint)

## Quick Start

### 1. Set Environment Variables

```bash
export OPENAI_API_KEY="your-api-key-here"
export OPENAI_MODEL="gpt-4"  # Optional, defaults to gpt-4
export OPENAI_URL="https://api.openai.com/v1"  # Optional, for Azure use your endpoint
```

### 2. Install

```bash
# Build from source
go mod download
go build -o gh-commit-src .

# Or install as GitHub CLI extension
gh extension install ellisvalentiner/gh-commit-src
```

### 3. Use

```bash
# Generate commit message from staged changes
gh commit

# Or use as Git alias
git config --global alias.auto-commit '!gh commit'
git auto-commit
```

## Installation

### GitHub CLI Extension

```bash
gh extension install ellisvalentiner/gh-commit-src
```

### From Source

```bash
git clone https://github.com/ellisvalentiner/gh-commit-src.git
cd gh-commit-src
go build -o gh-commit-src .
sudo mv gh-commit-src /usr/local/bin/
```

### Git Alias Setup

After installation, set up a Git alias for convenience:

```bash
git config --global alias.auto-commit '!gh commit'
```

## Configuration

Configuration can be provided via a YAML config file or environment variables. Environment variables take precedence over config file values.

### Config File

Create a config file in one of these locations (searched in order):

1. Current directory: `.gh-commit-src.yaml` or `gh-commit-src.yaml`
2. Home directory: `~/.gh-commit-src.yaml`
3. XDG config directory: `~/.config/gh-commit-src/config.yaml`

Example config file (`gh-commit-src.yaml.example`):

```yaml
openai:
  api_key: "sk-your-api-key-here"
  url: "https://api.openai.com/v1"
  model: "gpt-4"
  api_version: "2024-12-01-preview"  # For Azure OpenAI

fine_tune:
  temperature: 0.7
  max_tokens: 500
  top_p: 0.9
  frequency_penalty: 0.0
  presence_penalty: 0.0

prompt:
  # override: "Custom prompt template"
  suffix: "Commit message as follows:"

code_block:
  patterns:
    - "```go"
    - "```python"
    - "```"
```

Copy the example file to get started:

```bash
cp gh-commit-src.yaml.example ~/.gh-commit-src.yaml
# Edit ~/.gh-commit-src.yaml with your settings
```

### Environment Variables

Environment variables override config file values. Useful for per-command customization.

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `OPENAI_API_KEY` | Your OpenAI API key | Yes* | - |
| `OPENAI_URL` | API endpoint URL | No | `https://api.openai.com/v1` |
| `OPENAI_MODEL` | Model to use | No | `gpt-4` |
| `FINE_TUNE_PARAMS` | JSON parameters for model tuning | No | `{}` |
| `PROMPT_OVERRIDE` | Custom prompt template | No | Default prompt |
| `AZURE_API_VERSION` | Azure API version | No | `2024-12-01-preview` |
| `COMMIT_MESSAGE_SUFFIX` | Suffix for commit message | No | `Commit message as follows:` |
| `CODE_BLOCK_PATTERNS` | Comma-separated code block patterns | No | Default patterns |

\* Required if not set in config file

### Fine-Tune Parameters

Configure model behavior using the `fine_tune` section in the config file or `FINE_TUNE_PARAMS` environment variable:

**Config file:**

```yaml
fine_tune:
  temperature: 0.7
  max_tokens: 500
  top_p: 0.9
```

**Environment variable:**

```bash
export FINE_TUNE_PARAMS='{"temperature": 0.7, "max_tokens": 500, "top_p": 0.9}'
```

Supported parameters:

- `temperature` (float): Controls randomness (0.0-2.0)
- `max_tokens` (int): Maximum tokens in response
- `top_p` (float): Nucleus sampling parameter (0.0-1.0)
- `frequency_penalty` (float): Reduce repetition (-2.0 to 2.0)
- `presence_penalty` (float): Encourage new topics (-2.0 to 2.0)

### Quick Setup Examples

**Using config file:**

```bash
# Create config file
cp gh-commit-src.yaml.example ~/.gh-commit-src.yaml
# Edit with your API key
nano ~/.gh-commit-src.yaml
```

**Using environment variables:**

```bash
export OPENAI_API_KEY="sk-..."
export OPENAI_MODEL="gpt-4"
export FINE_TUNE_PARAMS='{"temperature": 0.7}'
```

**Inline (one-time use):**

```bash
OPENAI_API_KEY="sk-..." OPENAI_MODEL="gpt-4" gh commit
```

## Usage

### Basic Usage

```bash
# Stage your changes
git add .

# Generate and view commit message
gh commit

# Or commit directly with generated message
git commit -m "$(gh commit)"
```

### Advanced Usage

```bash
# Get commit statistics
gh commit --stats

# Ask a question (uses the same LLM)
gh commit --ask "What is the main change in this diff?"
```

### Commit Message Format

Generated commit messages follow the Conventional Commits specification:

```md
<type>(<scope>): <subject>

<body>

<footer>
```

Types include: `fix`, `feat`, `perf`, `revert`, etc.

## Examples

### Example Output

```md

feat(util): add configuration support for fine-tuning parameters

- Implement FINE_TUNE_PARAMS environment variable parsing
- Add support for temperature, max_tokens, top_p, and penalty parameters
- Update documentation with configuration examples
- Add validation for JSON parameter format

Closes #123
```

## Development

### Building

```bash
# Download dependencies
go mod download

# Build
go build -v .

# Run tests
go test -v ./...

# Format code
go fmt ./...

# Run linter
golangci-lint run
```

### Project Structure

```text
.
├── main.go           # CLI entry point
├── util.go           # Core functionality
├── util_test.go      # Tests
├── go.mod            # Go dependencies
└── .github/          # GitHub workflows and configs
    ├── workflows/    # CI/CD workflows
    └── labeler.yml   # PR labeling rules
```

## Contributing

Contributions are welcome! Please follow these steps:

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes** and ensure tests pass
4. **Commit your changes**: `git commit -m 'feat: add amazing feature'`
5. **Push to the branch**: `git push origin feature/amazing-feature`
6. **Open a Pull Request**

### Development Setup

```bash
# Clone your fork
git clone https://github.com/your-username/gh-commit-src.git
cd gh-commit-src

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -v .
```

### Code Quality

This project uses:

- `golangci-lint` for linting
- `go vet` for static analysis
- `go fmt` for code formatting
- GitHub Actions for CI/CD

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Original project: [github.com/megamanics/gh-commit](https://github.com/megamanics/gh-commit)
- OpenAI for providing the language models
- All contributors who have helped improve this project

## Related Projects

- [Conventional Commits](https://www.conventionalcommits.org/) - Commit message specification
- [GitHub CLI](https://cli.github.com/) - Official GitHub CLI tool
