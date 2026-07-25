# Claude Desktop Configuration

Configure Claude Desktop to use the Atlassian MCP Server.

## Configuration File Location

| OS | Path |
|----|------|
| macOS | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| Windows | `%APPDATA%\Claude\claude_desktop_config.json` |
| Linux | `~/.config/Claude/claude_desktop_config.json` |

## Basic Configuration

### With Direct Credentials

```json
{
  "mcpServers": {
    "atlassian": {
      "command": "/path/to/mcp-atlassian",
      "env": {
        "ATLASSIAN_URL": "https://example.atlassian.net",
        "ATLASSIAN_USERNAME": "user@example.com",
        "ATLASSIAN_API_TOKEN": "your-api-token"
      }
    }
  }
}
```

### With 1Password

```json
{
  "mcpServers": {
    "atlassian": {
      "command": "/path/to/mcp-atlassian",
      "env": {
        "ATLASSIAN_URL": "https://example.atlassian.net",
        "OP_SERVICE_ACCOUNT_TOKEN": "ops_...",
        "OMNITOKEN_VAULT_URI": "op://MyVault",
        "OMNITOKEN_CREDENTIALS_NAME": "atlassian"
      }
    }
  }
}
```

### With Bitwarden

```json
{
  "mcpServers": {
    "atlassian": {
      "command": "/path/to/mcp-atlassian",
      "env": {
        "ATLASSIAN_URL": "https://example.atlassian.net",
        "BW_ACCESS_TOKEN": "...",
        "BW_ORGANIZATION_ID": "...",
        "OMNITOKEN_VAULT_URI": "bw://org-id",
        "OMNITOKEN_CREDENTIALS_NAME": "atlassian"
      }
    }
  }
}
```

### With Keeper

```json
{
  "mcpServers": {
    "atlassian": {
      "command": "/path/to/mcp-atlassian",
      "env": {
        "ATLASSIAN_URL": "https://example.atlassian.net",
        "KSM_TOKEN": "US:...",
        "OMNITOKEN_VAULT_URI": "keeper://",
        "OMNITOKEN_CREDENTIALS_NAME": "atlassian"
      }
    }
  }
}
```

## Environment Variables Reference

| Variable | Description |
|----------|-------------|
| `ATLASSIAN_URL` | Atlassian instance URL |
| `ATLASSIAN_USERNAME` | Atlassian username/email |
| `ATLASSIAN_API_TOKEN` | Atlassian API token |
| `OMNITOKEN_VAULT_URI` | Vault URI (e.g., `op://MyVault`) |
| `OMNITOKEN_CREDENTIALS_NAME` | Credential name in vault (default: `confluence`) |
| `OP_SERVICE_ACCOUNT_TOKEN` | 1Password service account token |
| `BW_ACCESS_TOKEN` | Bitwarden access token |
| `BW_ORGANIZATION_ID` | Bitwarden organization ID |
| `KSM_TOKEN` | Keeper token (format: `REGION:TOKEN`) |

## Multiple Servers

You can run multiple MCP servers alongside Atlassian:

```json
{
  "mcpServers": {
    "atlassian": {
      "command": "/path/to/mcp-atlassian",
      "env": {
        "ATLASSIAN_URL": "https://example.atlassian.net",
        "ATLASSIAN_USERNAME": "user@example.com",
        "ATLASSIAN_API_TOKEN": "token"
      }
    },
    "google": {
      "command": "/path/to/mcp-google",
      "env": {
        "GOOGLE_CREDENTIALS_FILE": "/path/to/service-account.json"
      }
    },
    "aha": {
      "command": "/path/to/mcp-aha",
      "env": {
        "AHA_DOMAIN": "mycompany",
        "AHA_API_TOKEN": "token"
      }
    }
  }
}
```

## Troubleshooting

### Server Not Starting

Check the Claude Desktop logs:

- macOS: `~/Library/Logs/Claude/`
- Windows: `%APPDATA%\Claude\logs\`

Common issues:

1. **Binary not found**: Verify the path is correct
2. **Credentials not found**: Check environment variables
3. **Permission denied**: Ensure the binary is executable (`chmod +x`)
4. **Invalid URL**: Check `ATLASSIAN_URL` format

### Verifying Configuration

Test the server manually:

```bash
# Should start and wait for input (Ctrl+C to exit)
/path/to/mcp-atlassian --base-url https://example.atlassian.net \
                        --username user@example.com \
                        --api-token your-token
```

### JSON Syntax Errors

Validate your JSON:

```bash
# On macOS/Linux
cat ~/Library/Application\ Support/Claude/claude_desktop_config.json | python3 -m json.tool
```

## Available Tools in Claude

Once configured, you can ask Claude to:

- "Read Confluence page 12345"
- "Search for pages about authentication"
- "Create a new page in TEAM space"
- "Update page 12345 with a status table"
- "Delete page 67890"
