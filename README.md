# RuleForge

A CLI tool for managing AI agent rules. Synchronizes AI agent rule files (such as `.cursor/rules.md`) between a base repository and local repositories.

## Overview

This tool provides the following functions:

1. **Download**: Copy rule files from the base repository to the current directory
2. **Upload**: Send rule files from the current directory to the base repository as a PR
3. **Update Notification**: Automatically checks for new versions and notifies when updates are available

## Installation

```bash
go install github.com/yourusername/ruleforge@latest
```

Or clone the repository and build manually:

```bash
git clone https://github.com/yourusername/ruleforge.git
cd ruleforge
go build
```

## Usage

### Basic Commands

```bash
# Display help
ruleforge --help

# Download rules from the base repository
ruleforge download --base-repo https://github.com/organization/base-rules-repo

# Upload rules from the current directory to the base repository as a PR
ruleforge upload --base-repo https://github.com/organization/base-rules-repo --message "Update rules for my-project"
```

### Configuration File

Create a `.ruleforge.yaml` configuration file to omit command line arguments:

```yaml
base-repo: https://github.com/organization/base-rules-repo
target-files:
  - .cursor/rules.md
  - .cursor/config.json
github-token: ${GITHUB_TOKEN} # Load from environment variable
```

## Architecture

```
cmd/             # Entry points and CLI command definitions
  ruleforge/
    main.go
internal/        # Internal packages
  config/        # Configuration file related
  download/      # Download functionality
  upload/        # Upload functionality
  github/        # GitHub API operations
  file/          # File operation utilities
  logger/        # Logging
pkg/             # Public API packages (if needed)
```

## Development

### Requirements

- Go 1.20 or higher
- GitHub Personal Access Token (used for upload functionality)
- GoReleaser (for creating releases)

### Testing

```bash
make test
```

### Building

```bash
make build
```

### Releasing a New Version

To create a new release:

1. Make sure all your changes are committed and pushed
2. Run the release command with the new version:

```bash
make release VERSION=v1.2.3
```

This will:

- Create a git tag for the specified version
- Push the tag to GitHub
- GitHub Actions will automatically:
  - Build binaries for different platforms (Linux, macOS, Windows)
  - Create a GitHub Release
  - Upload the binaries to the GitHub release

The release workflow is defined in `.github/workflows/release.yml` and is triggered automatically when a tag with the format `v*` is pushed.

To test the release process without actually publishing:

```bash
make release-dry-run
```

## Available Make Commands

The project includes a Makefile with the following commands:

- `make build` - Build the binary
- `make test` - Run tests
- `make lint` - Run linter
- `make vet` - Run Go vet
- `make format` - Format code
- `make clean` - Clean build artifacts
- `make release` - Create a new release
- `make release-dry-run` - Test the release process
- `make install` - Install locally

## License

MIT

## Contributing

1. Fork this repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push the branch (`git push origin feature/amazing-feature`)
5. Create a Pull Request
