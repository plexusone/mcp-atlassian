# Environment Variables

All command-line flags can be set via environment variables.

## Available Variables

### Credential Configuration

| Variable | Flag | Description |
|----------|------|-------------|
| `ATLASSIAN_URL` | `--base-url` | Atlassian instance URL |
| `ATLASSIAN_USERNAME` | `--username` | Atlassian username/email |
| `ATLASSIAN_API_TOKEN` | `--api-token` | Atlassian API token |
| `OMNITOKEN_VAULT_URI` | `--vault` | Vault URI for credentials |
| `OMNITOKEN_CREDENTIALS_NAME` | `--credentials-name` | Name of credentials in vault |

### Vault Provider Authentication

| Variable | Description |
|----------|-------------|
| `OP_SERVICE_ACCOUNT_TOKEN` | 1Password service account token |
| `BW_ACCESS_TOKEN` | Bitwarden access token |
| `BW_ORGANIZATION_ID` | Bitwarden organization ID |
| `KSM_TOKEN` | Keeper token (format: `REGION:TOKEN`) |
| `KSM_CONFIG` | Keeper config (base64-encoded JSON) |

## Precedence

Command-line flags take precedence over environment variables.

```bash
# Environment variable is used
export ATLASSIAN_URL=https://example.atlassian.net
mcp-atlassian
# Uses: https://example.atlassian.net

# Flag overrides environment
export ATLASSIAN_URL=https://example.atlassian.net
mcp-atlassian --base-url https://other.atlassian.net
# Uses: https://other.atlassian.net
```

## Examples

### Direct Credentials

```bash
export ATLASSIAN_URL="https://example.atlassian.net"
export ATLASSIAN_USERNAME="user@example.com"
export ATLASSIAN_API_TOKEN="your-api-token"
mcp-atlassian
```

### 1Password

```bash
export OP_SERVICE_ACCOUNT_TOKEN="ops_..."
export OMNITOKEN_VAULT_URI=op://MyVault
export OMNITOKEN_CREDENTIALS_NAME=confluence
export ATLASSIAN_URL="https://example.atlassian.net"
mcp-atlassian
```

### Bitwarden

```bash
export BW_ACCESS_TOKEN="..."
export BW_ORGANIZATION_ID="..."
export OMNITOKEN_VAULT_URI=bw://org-id
export OMNITOKEN_CREDENTIALS_NAME=confluence
export ATLASSIAN_URL="https://example.atlassian.net"
mcp-atlassian
```

### Keeper

```bash
export KSM_TOKEN="US:..."
export OMNITOKEN_VAULT_URI=keeper://
export OMNITOKEN_CREDENTIALS_NAME=confluence
export ATLASSIAN_URL="https://example.atlassian.net"
mcp-atlassian
```

## Shell Configuration

### Bash/Zsh

Add to `~/.bashrc` or `~/.zshrc`:

```bash
# Atlassian MCP Server credentials
export ATLASSIAN_URL="https://example.atlassian.net"
export ATLASSIAN_USERNAME="user@example.com"
export ATLASSIAN_API_TOKEN="your-api-token"
```

### Fish

Add to `~/.config/fish/config.fish`:

```fish
set -gx ATLASSIAN_URL "https://example.atlassian.net"
set -gx ATLASSIAN_USERNAME "user@example.com"
set -gx ATLASSIAN_API_TOKEN "your-api-token"
```
