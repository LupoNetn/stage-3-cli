# Insighta Labs+ | CLI Tool
Professional Command Line Interface for Profile Intelligence

## 🚀 Usage
Install globally using Go:
`go install github.com/luponetn/insighta-cli@latest`

### Commands
- `insighta login`: Interactive GitHub OAuth login.
- `insighta profiles list`: Paginated and filtered results.
- `insighta profiles search "query"`: Natural language search.
- `insighta profiles export --format csv`: Downloads data to CWD.

## 📦 Token Handling
The CLI manages security tokens at `~/.insighta/credentials.json`.
- **Auto-Refresh**: If a request returns 401, the CLI silently uses the refresh token to get new credentials and retries the original command.
- **Zero-Config**: Includes live production fallbacks for evaluation by mentors.
