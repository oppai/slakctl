# slakctl

A command-line tool for managing Slack workspaces using personal tokens or OAuth2 authentication.

## Features

- **Authentication**: OAuth2 flow with automatic browser opening or manual token entry
- **Search**: Cross-channel message search functionality
- **Channel Management**: List all accessible channels in your workspace
- **Message Posting**: Send messages to specific channels

## Installation

### Prerequisites

- Go 1.24.3 or later
- [mise](https://github.com/jdx/mise) for task running

### Build from Source

```bash
git clone https://github.com/yourusername/slakctl
cd slakctl
mise run build
```

## Setup

### Option 1: OAuth2 Authentication (Recommended)

OAuth2 authentication automatically opens a browser and handles the token exchange process.

#### 1. Create a Slack App

1. Go to [Slack API](https://api.slack.com/apps)
2. Click "Create New App" → "From scratch"
3. Enter an app name and select your workspace
4. Navigate to "OAuth & Permissions"
5. **IMPORTANT**: In the "Redirect URLs" section, click "Add New Redirect URL" and add:
   ```
   https://b1c3-42-148-67-84.ngrok-free.app/callback
   ```
   Then click "Save URLs"
6. Add the following scopes under "Bot Token Scopes":
   - `channels:history` - View messages and other content in a user's public channels
   - `channels:read` - View basic information about public channels in a workspace
   - `channels:write` - Manage a user's public channels and create new ones on a user's behalf
   - `chat:write` - Send messages on a user's behalf
   - `search:read` - Search a workspace's content
7. Note down your **Client ID** and **Client Secret** from "Basic Information"

#### 2. Configure slakctl

Set your Slack app credentials:

```bash
./bin/slakctl config set
# Enter your Client ID and Client Secret when prompted
```

#### 3. Authenticate

Start the OAuth2 flow (opens browser automatically):

```bash
./bin/slakctl config oauth
```

### Option 2: Manual Token Authentication

If you prefer to use a personal token directly:

#### 1. Get a Slack Token

1. Go to [Slack API](https://api.slack.com/apps)
2. Create a new app or select an existing one
3. Navigate to "OAuth & Permissions"
4. Add the required scopes:
   - `channels:history` - View messages and other content in a user's public channels
   - `channels:read` - View basic information about public channels in a workspace
   - `channels:write` - Manage a user's public channels and create new ones on a user's behalf
   - `chat:write` - Send messages on a user's behalf
   - `search:read` - Search a workspace's content
5. Install the app to your workspace
6. Copy the "Bot User OAuth Token" (starts with `xoxb-`)

#### 2. Authenticate

```bash
./bin/slakctl auth token <your-token>
```

Or run without arguments to be prompted for the token:

```bash
./bin/slakctl auth token
```

## Usage

### Configuration Management

Check your current configuration:

```bash
slakctl config show
```

Set up OAuth2 credentials:

```bash
slakctl config set
```

### Authentication

OAuth2 authentication (recommended):

```bash
slakctl config oauth
```

Or manual token authentication:

```bash
slakctl auth token xoxb-your-token-here
```

### Search Messages

Search for messages across all channels:

```bash
slakctl search "deployment"
slakctl search "bug report"
```

### List Channels

Get a list of all channels you have access to:

```bash
slakctl channel list
```

### Post Messages

Send a message to a specific channel:

```bash
slakctl post "#general" "Hello, world!"
slakctl post "general" "Hello without # prefix"
```

## Configuration

slakctl stores your configuration in `~/.slakctl` as a JSON file. The file contains:

```json
{
  "token": "your-slack-token"
}
```

## Security

- Configuration file is created with 600 permissions (readable only by owner)
- Tokens are stored locally and never transmitted except to Slack's API
- Always use Bot User OAuth Tokens, not legacy tokens

## Development

### Available Tasks

View all available tasks:

```bash
mise tasks
```

### Common Development Tasks

**Build the project:**
```bash
mise run build
```

**Run tests:**
```bash
mise run test
```

**Run tests with verbose output:**
```bash
mise run test-verbose
```

**Format and lint code:**
```bash
mise run lint
```

**Run all checks (lint + test):**
```bash
mise run check
```

**Development workflow (format, lint, test, build):**
```bash
mise run dev
```

**Clean build artifacts:**
```bash
mise run clean
```

**Install to $GOPATH/bin:**
```bash
mise run install
```

### Project Structure

```
slakctl/
    bin/                 # Built binary location
    cmd/                 # Command implementations
        auth.go         # Authentication command
        channel.go      # Channel management commands
        config.go       # Configuration management commands
        post.go         # Message posting command
        root.go         # Root command and CLI setup
        search.go       # Search command
    internal/
        auth/           # OAuth2 authentication
            oauth.go
        config/         # Configuration management
            config.go
        slack/          # Slack API client
            client.go
    main.go             # Application entry point
    go.mod              # Go module file
```

## API Reference

### Commands

#### `slakctl config set`

Set up Slack app credentials for OAuth2 authentication.

**Example:**
```bash
slakctl config set
```

#### `slakctl config show`

Show current configuration (credentials are masked for security).

**Example:**
```bash
slakctl config show
```

#### `slakctl config oauth`

Authenticate using OAuth2 flow (opens browser automatically).

**Example:**
```bash
slakctl config oauth
```

#### `slakctl auth token [token]`

Authenticate with Slack using a personal token.

**Arguments:**
- `token` (optional): Slack token. If not provided, you'll be prompted to enter it.

**Example:**
```bash
slakctl auth token xoxb-123456789-abcdef
```

#### `slakctl search <keyword>`

Search for messages containing the specified keyword across all channels.

**Arguments:**
- `keyword` (required): The search term to look for in messages.

**Example:**
```bash
slakctl search "deployment failed"
```

#### `slakctl channel list`

List all channels in the workspace that you have access to.

**Example:**
```bash
slakctl channel list
```

#### `slakctl post <channel> <message>`

Post a message to the specified channel.

**Arguments:**
- `channel` (required): Channel name (with or without # prefix)
- `message` (required): Message text to send

**Example:**
```bash
slakctl post "#general" "Hello, team!"
slakctl post "general" "Hello without # prefix"
```

## Error Handling

Common errors and solutions:

### OAuth2 Errors

- **"redirect_uri did not match any configured URIs"**: 
  1. Go to your Slack App settings at [api.slack.com/apps](https://api.slack.com/apps)
  2. Select your app and navigate to "OAuth & Permissions"
  3. In the "Redirect URLs" section, ensure `https://b1c3-42-148-67-84.ngrok-free.app/callback` is added
  4. Click "Save URLs" after adding the redirect URL
  5. Try the OAuth flow again

- **"invalid_client_id"**: Double-check your Client ID in `slakctl config show`

- **"invalid_client"**: Double-check your Client Secret (re-run `slakctl config set` if needed)

### Authentication Errors

- **"no authentication token found"**: Run `slakctl auth token` or `slakctl config oauth` first
- **"authentication failed"**: Check if your token is valid and has the required scopes

### Permission Errors

- **"channel_not_found"**: Make sure the channel exists and you have access to it
- **"not_in_channel"**: You need to be a member of the channel to post messages

### Network Errors

- Check your internet connection
- Verify that you can access slack.com
- Check if there are any firewall restrictions

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Cobra](https://github.com/spf13/cobra) for the excellent CLI framework
- [Slack API](https://api.slack.com/) for providing comprehensive API documentation
- [mise](https://github.com/jdx/mise) for task running
