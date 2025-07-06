# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

This project uses `mise` for task management. All development commands should be run using mise:

```bash
# Build the project
mise run build

# Run tests
mise run test

# Run tests with verbose output
mise run test-verbose

# Run a single test file
go test ./cmd -v -run TestSpecificFunction

# Format and lint code
mise run lint

# Run all checks (lint + test)
mise run check

# Full development workflow (format, lint, test, build)
mise run dev

# Clean build artifacts
mise run clean
```

## Architecture Overview

slakctl is a CLI tool for managing Slack workspaces built with Go and Cobra. The architecture follows a clean separation of concerns:

### Command Structure
- **cmd/**: Contains all CLI command implementations using Cobra framework
  - `root.go`: Main command entry point, registers all subcommands
  - `auth.go`: Manual token authentication
  - `config.go`: OAuth2 credential management and OAuth flow
  - `search.go`: Message search with count/format options
  - `channel.go`: Channel listing functionality
  - `post.go`: Message posting to channels

### Core Components
- **internal/config/**: Configuration management
  - Handles `~/.slakctl` JSON file with 600 permissions
  - Stores OAuth2 credentials (ClientID, ClientSecret) and tokens
  - Auto-creates empty config if file doesn't exist

- **internal/slack/**: Slack API client
  - `Client` struct with token-based authentication
  - `makeRequest()` handles all HTTP communication with Slack API
  - Supports search with count limits, channel listing, message posting
  - Search supports custom formatting with template variables

- **internal/auth/**: OAuth2 authentication flow
  - Handles browser-based OAuth2 flow
  - Uses hardcoded ngrok redirect URL for callback

### Authentication Flow
The application supports two authentication methods:
1. **OAuth2 (recommended)**: Browser-based flow requiring Client ID/Secret
2. **Manual token**: Direct token entry for Bot User OAuth Tokens

### Configuration Management
- Configuration stored in `~/.slakctl` JSON file
- Contains: `token`, `client_id`, `client_secret`
- `config show` command masks sensitive data for security
- Uses `cmd.Printf()` instead of `fmt.Printf()` for testable output

### Search Functionality
- Supports count limiting (1-100 messages, default 20)
- Multiple output formats: text, json, custom templates
- Template variables: `{channel}`, `{user}`, `{text}`, `{timestamp}`, `{permalink}`, etc.
- Search results include pagination info and total count

### Error Handling Patterns
- Functions that access command arguments must validate `len(args)` before array access
- All command functions should return descriptive error messages
- Use `fmt.Errorf()` with `%w` for error wrapping
- Network errors and API responses are properly handled with status code checking

### Testing Notes
- Tests create temporary directories and set `HOME` environment variable
- Use `cmd.SetOut()` and `cmd.SetErr()` for output capture in tests
- Command functions should use `cmd.Printf()` instead of `fmt.Printf()` for testability
- Tests validate both success and error scenarios
- **Important Development Practice**:
  - 機能やロジックを追加したときは必ずテストを追加する

### Required Slack API Scopes
- `channels:history`: View message content
- `channels:read`: View channel information
- `channels:write`: Manage channels
- `chat:write`: Send messages
- `search:read`: Search workspace content