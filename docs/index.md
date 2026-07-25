# Atlassian MCP Server

An MCP server for Atlassian products (Confluence and Jira) built on the [omniskill](https://github.com/plexusone/omniskill) framework.

## Features

- **26 MCP tools** across Confluence and Jira
- **Confluence**: Safe handling of Storage Format XHTML with structured content blocks
- **Jira**: Search, create, update, clone, bulk operations, agile reports
- **Shared Atlassian auth** - One set of credentials for both products
- **Vault-backed credentials** - 1Password, Bitwarden, Keeper support
- **OAuth 2.1 with PKCE** - Authorization code flow for multi-user deployments

## Available Tools

### Confluence Tools

| Tool | Description |
|------|-------------|
| `confluence_read_page` | Read a page as structured blocks |
| `confluence_read_page_xhtml` | Read a page as raw Storage Format XHTML |
| `confluence_update_page` | Update a page with structured blocks |
| `confluence_update_page_xhtml` | Update a page with raw Storage Format XHTML |
| `confluence_create_page` | Create a new page with structured blocks |
| `confluence_create_table` | Create a table block from structured data |
| `confluence_delete_page` | Delete a page |
| `confluence_search_pages` | Search pages using CQL |

### Jira Tools

| Tool | Description |
|------|-------------|
| `jira_get_issue` | Get a Jira issue by key |
| `jira_search` | Search issues using JQL |
| `jira_create_issue` | Create a new issue |
| `jira_update_issue` | Update issue fields |
| `jira_add_comment` | Add a comment to an issue |
| `jira_get_comments` | Get comments for an issue |
| `jira_get_transitions` | Get available transitions |
| `jira_transition_issue` | Transition an issue to a new status |
| `jira_clone_issue` | Clone an issue with field mapping |
| `jira_bulk_update` | Update multiple issues at once |
| `jira_get_projects` | List available projects |
| `jira_get_boards` | List agile boards |
| `jira_get_sprints` | List sprints for a board |
| `jira_move_to_sprint` | Move issues to a sprint |
| `jira_velocity_report` | Sprint velocity report |
| `jira_burndown_report` | Sprint burndown data |
| `jira_worklog_report` | Time tracking summary |
| `jira_cycle_time_report` | Issue cycle time analysis |

## Quick Start

```bash
# Install
go install github.com/plexusone/mcp-atlassian/cmd/mcp-atlassian@latest

# Configure credentials
export ATLASSIAN_URL="https://example.atlassian.net"
export ATLASSIAN_USERNAME="user@example.com"
export ATLASSIAN_API_TOKEN="your-api-token"

# Run as MCP server
mcp-atlassian
```

## Next Steps

- [Installation](getting-started/installation.md) - Install the server
- [Setup](getting-started/setup.md) - Configure your credentials
- [Quick Start](getting-started/quickstart.md) - Start using the tools
- [Tools Reference](tools/overview.md) - Detailed tool documentation
- [Storage Format](storage-format/overview.md) - Understanding block types
