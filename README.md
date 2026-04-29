# Insighta Labs+ | CLI Tool
Professional Command Line Interface for Profile Intelligence

## 🚀 Usage
Install globally using Go:
`go install github.com/luponetn/insighta-cli@latest`

### Commands
- **Authentication**:
  - `insighta login`: Interactive GitHub OAuth login with PKCE.
  - `insighta logout`: Invalidates session on server and clears local tokens.
  - `insighta whoami`: Displays the currently authenticated identity.
- **Profiles**:
  - `insighta profiles list`: Paginated and filtered results (supports flags like `--gender`, `--country`, `--min-age`).
  - `insighta profiles search "query"`: Natural language search (e.g. "young males from nigeria").
  - `insighta profiles get <id>`: Fetch detailed profile data.
  - `insighta profiles create --name "Name"`: Register a new profile (Admin only).
  - `insighta profiles export --format csv`: Downloads filtered data to your current directory.

## 📦 Token Handling
The CLI manages security tokens at `~/.insighta/credentials.json`.
- **Auto-Refresh**: Implements silent token refresh. If a request returns `401 Unauthorized`, the CLI automatically exchanges the refresh token for a new pair and retries the request.
- **Zero-Config**: Uses environment variables from `.env` if present, falling back to production endpoints for a seamless mentor evaluation experience.
