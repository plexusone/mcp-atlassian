# MCP Server for Atlassian

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Docs][docs-mkdoc-svg]][docs-mkdoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/mcp-atlassian/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/mcp-atlassian/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/mcp-atlassian/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/mcp-atlassian/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/mcp-atlassian/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/mcp-atlassian/actions/workflows/go-sast-codeql.yaml
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/mcp-atlassian
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/mcp-atlassian
 [docs-mkdoc-svg]: https://img.shields.io/badge/Go-dev%20guide-blue.svg
 [docs-mkdoc-url]: https://plexusone.dev/mcp-atlassian
 [viz-svg]: https://img.shields.io/badge/Go-visualizaton-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=plexusone%2Fmcp-atlassian
 [loc-svg]: https://tokei.rs/b1/github/plexusone/mcp-atlassian
 [repo-url]: https://github.com/plexusone/mcp-atlassian
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/mcp-atlassian/blob/main/LICENSE

An MCP server for Atlassian products (Confluence and Jira) built on the [omniskill](https://github.com/plexusone/omniskill) framework. Provides 26 tools for AI assistants to interact with Confluence pages and Jira issues.

## Features

- **Confluence tools** — Read, create, update, and search Confluence pages with safe handling of Storage Format (XHTML)
- **Jira tools** — Search, create, update, clone, and transition issues; agile boards, sprints, and reporting
- **Shared Atlassian auth** — Single `ATLASSIAN_URL` credential for both products
- **OAuth 2.1 with PKCE** — Serve over HTTP with authorization code flow and optional ngrok tunneling
- **Vault-backed credentials** — 1Password, Bitwarden, Keeper support via [omnitoken](https://github.com/plexusone/omnitoken)
- **Composable skills** — Each product is an independent omniskill that can be used standalone or combined

## Packages

| Package | Description |
|---------|-------------|
| `skills/confluence` | Confluence omniskill — 8 tools for page operations |
| `skills/jira` | Jira omniskill — 18 tools for issue tracking, agile, and reporting |
| `cmd/mcp-atlassian` | MCP server binary with stdio and HTTP modes |

The underlying Confluence and Jira client libraries live in [Go-Atlassian](https://github.com/grokify/go-atlassian) (`confluence/` and `jira/` packages).

## Installation

### MCP Server Binary

```bash
go install github.com/plexusone/mcp-atlassian/cmd/mcp-atlassian@latest
```

### As a Library

```bash
go get github.com/plexusone/mcp-atlassian
```

## Quick Start

### Using the Storage Package

```go
import "github.com/grokify/go-atlassian/confluence/storage"

// Create a page with structured blocks
page := &storage.Page{
    Blocks: []storage.Block{
        &storage.Heading{Level: 1, Text: "Welcome"},
        &storage.Paragraph{Text: "This is a test page."},
        &storage.Table{
            Headers: []string{"Name", "Status"},
            Rows: []storage.Row{
                {Cells: []storage.Cell{
                    {Text: "Service A"},
                    {Macro: &storage.Macro{
                        Name:   "status",
                        Params: map[string]string{"colour": "Green", "title": "OK"},
                    }},
                }},
            },
        },
    },
}

// Render to Storage XHTML
xhtml, err := storage.Render(page)
if err != nil {
    log.Fatal(err)
}

// Validate before sending to Confluence
if err := storage.Validate(xhtml); err != nil {
    log.Fatal(err)
}
```

### Using the Confluence Client

```go
import "github.com/grokify/go-atlassian/confluence"

// Create client
auth := confluence.BasicAuth{
    Username: "user@example.com",
    Token:    "your-api-token",
}
client := confluence.NewClient("https://example.atlassian.net/wiki", auth)

// Get a page as structured IR
page, info, err := client.GetPageStorage(ctx, "12345")
if err != nil {
    log.Fatal(err)
}

// Modify the page
page.Blocks = append(page.Blocks, &storage.Paragraph{Text: "New content"})

// Update the page
err = client.UpdatePageStorage(ctx, info.ID, page, info.Version, info.Title)
```

### Running the MCP Server

```bash
# Install from GitHub
go install github.com/plexusone/mcp-atlassian/cmd/mcp-atlassian@latest

# Or build from source
go build -o mcp-atlassian ./cmd/mcp-atlassian
```

### Configuring with Claude Code

Claude Code supports three configuration scopes. See [Claude Code MCP docs](https://code.claude.com/docs/en/mcp) for details.

**User scope** (`~/.claude.json`):

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

**Project scope** (`.mcp.json` in project root, can be checked into source control):

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

**Enterprise managed** (`managed-mcp.json` in system directories):

See [Enterprise MCP configuration](https://code.claude.com/docs/en/mcp) for details.

### Configuring with Claude Desktop

Add to your Claude Desktop settings (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

#### With Direct Credentials

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

#### With 1Password

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

#### With Bitwarden

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

#### With Keeper

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

### Option 2: Vault-Backed Credentials

Use [omnitoken](https://github.com/plexusone/omnitoken) with vault backends for secure credential storage.

| Provider | URI Pattern | Requirements |
|----------|-------------|--------------|
| 1Password | `op://vault` | `OP_SERVICE_ACCOUNT_TOKEN` env var |
| Bitwarden | `bw://org-id` | `BW_ACCESS_TOKEN` and `BW_ORGANIZATION_ID` env vars |
| Keeper | `keeper://` | `KSM_TOKEN` or `KSM_CONFIG` env var |
| File | `file:///path` | None |

#### 1Password Example

```bash
export OP_SERVICE_ACCOUNT_TOKEN="ops_..."
mcp-atlassian --vault op://MyVault --credentials-name atlassian \
               --base-url https://example.atlassian.net
```

#### Bitwarden Example

```bash
export BW_ACCESS_TOKEN="..."
export BW_ORGANIZATION_ID="..."
mcp-atlassian --vault bw://org-id --credentials-name atlassian \
               --base-url https://example.atlassian.net
```

#### Keeper Example

```bash
export KSM_TOKEN="US:..."
mcp-atlassian --vault keeper:// --credentials-name atlassian \
               --base-url https://example.atlassian.net
```

### Environment Variables

| Variable | Flag | Description |
|----------|------|-------------|
| `ATLASSIAN_URL` | `--base-url` | Your Atlassian instance URL (e.g., `https://example.atlassian.net`) |
| `ATLASSIAN_USERNAME` | `--username` | Your Atlassian username (usually your email) |
| `ATLASSIAN_API_TOKEN` | `--api-token` | API token from [Atlassian Account Settings](https://id.atlassian.com/manage-profile/security/api-tokens) |
| `CONFLUENCE_BASE_URL` | - | Legacy: Confluence URL with `/wiki` suffix (falls back if `ATLASSIAN_URL` not set) |
| `OMNITOKEN_VAULT_URI` | `--vault` | Vault URI for credentials |
| `OMNITOKEN_CREDENTIALS_NAME` | `--credentials-name` | Name of credentials in vault (default: `atlassian`) |
| `OP_SERVICE_ACCOUNT_TOKEN` | - | 1Password service account token |
| `BW_ACCESS_TOKEN` | - | Bitwarden access token |
| `BW_ORGANIZATION_ID` | - | Bitwarden organization ID |
| `KSM_TOKEN` | - | Keeper token (format: `REGION:TOKEN`) |

### Running Standalone (for testing)

```bash
export ATLASSIAN_URL=https://example.atlassian.net
export ATLASSIAN_USERNAME=user@example.com
export ATLASSIAN_API_TOKEN=your-api-token

./mcp-atlassian
```

## MCP Tools

The MCP server exposes these tools:

### Confluence

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

### Jira

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

### When to Use XHTML Tools

The structured block tools (`confluence_read_page`, `confluence_update_page`) are safer and recommended for most use cases. However, the XHTML tools are useful when:

- **Debugging**: See the raw XHTML to understand parsing issues
- **Complex content**: Tables with column widths, nested lists, or custom macros that the block parser doesn't fully support
- **Preserving formatting**: When you need to make small edits without losing inline styles or attributes

### Example Tool Inputs

#### confluence_read_page

```json
{
  "name": "confluence_read_page",
  "arguments": {
    "page_id": "12345"
  }
}
```

#### confluence_read_page_xhtml

```json
{
  "name": "confluence_read_page_xhtml",
  "arguments": {
    "page_id": "12345"
  }
}
```

Returns the raw Storage Format XHTML in the `xhtml` field, along with page metadata.

#### confluence_create_page

```json
{
  "name": "confluence_create_page",
  "arguments": {
    "space_key": "TEAM",
    "title": "Meeting Notes 2025-01-15",
    "parent_id": "11111",
    "blocks": [
      {"type": "heading", "level": 1, "text": "Meeting Notes"},
      {"type": "paragraph", "text": "Attendees: Alice, Bob, Carol"},
      {"type": "heading", "level": 2, "text": "Action Items"},
      {"type": "bullet_list", "items": ["Review PR #123", "Update documentation", "Schedule follow-up"]}
    ]
  }
}
```

#### confluence_update_page

```json
{
  "name": "confluence_update_page",
  "arguments": {
    "page_id": "12345",
    "title": "Updated Page Title",
    "blocks": [
      {"type": "heading", "level": 1, "text": "Updated Content"},
      {"type": "paragraph", "text": "This page has been updated."},
      {"type": "table", "headers": ["Name", "Role"], "rows": [["Alice", "Lead"], ["Bob", "Developer"]]}
    ]
  }
}
```

#### confluence_update_page_xhtml

```json
{
  "name": "confluence_update_page_xhtml",
  "arguments": {
    "page_id": "12345",
    "title": "Updated Page Title",
    "xhtml": "<h1>Updated Content</h1><p>This page has been updated with raw XHTML.</p><table><tbody><tr><th>Name</th><th>Role</th></tr><tr><td>Alice</td><td>Lead</td></tr></tbody></table>"
  }
}
```

Use this when you need to preserve complex formatting that would be lost with structured blocks.

#### confluence_create_table

```json
{
  "name": "confluence_create_table",
  "arguments": {
    "headers": ["Service", "Owner", "Status"],
    "rows": [
      ["Auth", "Platform", {"macro": {"name": "status", "params": {"colour": "Green", "title": "OK"}}}],
      ["API", "Backend", {"macro": {"name": "status", "params": {"colour": "Yellow", "title": "Degraded"}}}]
    ]
  }
}
```

#### confluence_delete_page

```json
{
  "name": "confluence_delete_page",
  "arguments": {
    "page_id": "12345"
  }
}
```

#### confluence_search_pages

```json
{
  "name": "confluence_search_pages",
  "arguments": {
    "cql": "space=TEAM and type=page and title~\"Meeting\"",
    "limit": 10
  }
}
```

## Block Types

| Type | Description |
|------|-------------|
| `Paragraph` | Text paragraph |
| `Heading` | H1-H6 headings |
| `Table` | Tables with headers, rows, and optional macros in cells |
| `BulletList` | Unordered list |
| `NumberedList` | Ordered list |
| `Macro` | Confluence macros (status, info, code, etc.) |
| `CodeBlock` | Code blocks with language |
| `HorizontalRule` | Horizontal divider |

## Why This Approach Works

1. **LLMs produce structured JSON** (not XHTML) → fewer errors
2. **Rendering is deterministic** Go code → guaranteed valid output
3. **Validation catches issues** before API calls
4. **Round-trip safe**: Parse → Modify → Render preserves structure

## License

MIT

## See Also

- [CHANGELOG.md](CHANGELOG.md) - Release history
- [ROADMAP.md](ROADMAP.md) - Planned features
- [Confluence Storage Format Documentation](https://confluence.atlassian.com/doc/confluence-storage-format-790796544.html)
