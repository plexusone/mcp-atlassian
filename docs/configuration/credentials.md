# Credentials

The Atlassian MCP Server supports multiple credential sources for authentication.

## Option 1: Direct Credentials

Provide your Confluence URL, username, and API token directly.

### Confluence Cloud

1. Go to [Atlassian Account Settings](https://id.atlassian.com/manage-profile/security/api-tokens)
2. Create a new API token
3. Use your email as the username

### Usage

```bash
mcp-atlassian --base-url https://example.atlassian.net \
               --username user@example.com \
               --api-token your-api-token
```

Or with environment variables:

```bash
export ATLASSIAN_URL="https://example.atlassian.net"
export ATLASSIAN_USERNAME="user@example.com"
export ATLASSIAN_API_TOKEN="your-api-token"
mcp-atlassian
```

## Option 2: Vault-Backed Credentials

Use [omnitoken](https://github.com/plexusone/omnitoken) with vault backends for secure credential storage.

**Note:** When using vault credentials, you still need to provide `--base-url` separately.

### Supported Vault URIs

| Provider | URI Pattern | Requirements |
|----------|-------------|--------------|
| 1Password | `op://vault` | `OP_SERVICE_ACCOUNT_TOKEN` env var |
| Bitwarden | `bw://org-id` | `BW_ACCESS_TOKEN` and `BW_ORGANIZATION_ID` env vars |
| Keeper | `keeper://` | `KSM_TOKEN` or `KSM_CONFIG` env var |
| File | `file:///path/to/dir` | None |

### 1Password

```bash
export OP_SERVICE_ACCOUNT_TOKEN="ops_..."
mcp-atlassian --vault op://MyVault --credentials-name confluence \
               --base-url https://example.atlassian.net
```

### Bitwarden

```bash
export BW_ACCESS_TOKEN="..."
export BW_ORGANIZATION_ID="..."
mcp-atlassian --vault bw://org-id --credentials-name confluence \
               --base-url https://example.atlassian.net
```

### Keeper

```bash
export KSM_TOKEN="US:..."
mcp-atlassian --vault keeper:// --credentials-name confluence \
               --base-url https://example.atlassian.net
```

### File Vault

For local development:

```bash
mcp-atlassian --vault file:///path/to/secrets --credentials-name confluence \
               --base-url https://example.atlassian.net
```

## Credential Format

When using vault storage, credentials should be in goauth format:

```json
{
  "type": "basic",
  "basic": {
    "username": "user@example.com",
    "password": "your-api-token",
    "serverURL": "https://example.atlassian.net"
  }
}
```

Or with OAuth2/Bearer token:

```json
{
  "type": "headerquery",
  "headerQuery": {
    "serverURL": "https://example.atlassian.net",
    "header": {
      "Authorization": ["Bearer your-token"]
    }
  }
}
```

## Security Best Practices

1. **Never commit credentials** - Add credentials files to `.gitignore`
2. **Use vault backends** - For production, use proper secrets management
3. **Rotate tokens** - Periodically rotate API tokens
4. **Limit scope** - Use tokens with minimum required permissions
