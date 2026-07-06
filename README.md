# MCP Server for Confluence

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Docs][docs-mkdoc-svg]][docs-mkdoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/mcp-confluence/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/mcp-confluence/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/mcp-confluence/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/mcp-confluence/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/mcp-confluence/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/mcp-confluence/actions/workflows/go-sast-codeql.yaml
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/mcp-confluence
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/mcp-confluence
 [docs-mkdoc-svg]: https://img.shields.io/badge/Go-dev%20guide-blue.svg
 [docs-mkdoc-url]: https://plexusone.dev/mcp-confluence
 [viz-svg]: https://img.shields.io/badge/Go-visualizaton-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=plexusone%2Fmcp-confluence
 [loc-svg]: https://tokei.rs/b1/github/plexusone/mcp-confluence
 [repo-url]: https://github.com/plexusone/mcp-confluence
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/mcp-confluence/blob/main/LICENSE

An MCP server for Confluence with safe handling of Confluence Storage Format (XHTML).

## The Problem

When AI assistants interact with Confluence via MCP servers, they often corrupt pages—especially tables—because:

1. LLMs generate Markdown or HTML5 instead of Confluence Storage XHTML
2. Tables require specific structure (`<tbody>`, no `<thead>`)
3. Macros need `ac:` namespaces
4. Round-tripping through incorrect formats causes data loss

## The Solution

This library provides:

- **Structured IR (Intermediate Representation)**: Work with Go types instead of raw XHTML
- **Safe Rendering**: IR → valid Storage XHTML with proper structure
- **Validation**: Catch forbidden tags, missing `<tbody>`, etc. before API calls
- **MCP Server**: Tools that accept structured JSON, never raw XHTML

## Packages

| Package | Description |
|---------|-------------|
| `storage` | IR types, render, parse, validate for Confluence Storage Format |
| `confluence` | REST API client with IR integration |
| `mcpserver` | MCP server implementation with structured tools |

## Installation

### MCP Server Binary

```bash
go install github.com/plexusone/mcp-confluence/cmd/mcp-confluence@latest
```

### As a Library

```bash
go get github.com/plexusone/mcp-confluence
```

## Quick Start

### Using the Storage Package

```go
import "github.com/plexusone/mcp-confluence/storage"

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
import "github.com/plexusone/mcp-confluence/confluence"

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
go install github.com/plexusone/mcp-confluence/cmd/mcp-confluence@latest

# Or build from source
go build -o mcp-confluence ./cmd/mcp-confluence
```

### Configuring with Claude Code

Claude Code supports three configuration scopes. See [Claude Code MCP docs](https://code.claude.com/docs/en/mcp) for details.

**User scope** (`~/.claude.json`):

```json
{
  "mcpServers": {
    "confluence": {
      "command": "/path/to/mcp-confluence",
      "env": {
        "CONFLUENCE_BASE_URL": "https://example.atlassian.net/wiki",
        "CONFLUENCE_USERNAME": "user@example.com",
        "CONFLUENCE_API_TOKEN": "your-api-token"
      }
    }
  }
}
```

**Project scope** (`.mcp.json` in project root, can be checked into source control):

```json
{
  "mcpServers": {
    "confluence": {
      "command": "/path/to/mcp-confluence",
      "env": {
        "CONFLUENCE_BASE_URL": "https://example.atlassian.net/wiki",
        "CONFLUENCE_USERNAME": "user@example.com",
        "CONFLUENCE_API_TOKEN": "your-api-token"
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
    "confluence": {
      "command": "/path/to/mcp-confluence",
      "env": {
        "CONFLUENCE_BASE_URL": "https://example.atlassian.net/wiki",
        "CONFLUENCE_USERNAME": "user@example.com",
        "CONFLUENCE_API_TOKEN": "your-api-token"
      }
    }
  }
}
```

#### With 1Password

```json
{
  "mcpServers": {
    "confluence": {
      "command": "/path/to/mcp-confluence",
      "env": {
        "CONFLUENCE_BASE_URL": "https://example.atlassian.net/wiki",
        "OP_SERVICE_ACCOUNT_TOKEN": "ops_...",
        "OMNITOKEN_VAULT_URI": "op://MyVault",
        "OMNITOKEN_CREDENTIALS_NAME": "confluence"
      }
    }
  }
}
```

#### With Bitwarden

```json
{
  "mcpServers": {
    "confluence": {
      "command": "/path/to/mcp-confluence",
      "env": {
        "CONFLUENCE_BASE_URL": "https://example.atlassian.net/wiki",
        "BW_ACCESS_TOKEN": "...",
        "BW_ORGANIZATION_ID": "...",
        "OMNITOKEN_VAULT_URI": "bw://org-id",
        "OMNITOKEN_CREDENTIALS_NAME": "confluence"
      }
    }
  }
}
```

#### With Keeper

```json
{
  "mcpServers": {
    "confluence": {
      "command": "/path/to/mcp-confluence",
      "env": {
        "CONFLUENCE_BASE_URL": "https://example.atlassian.net/wiki",
        "KSM_TOKEN": "US:...",
        "OMNITOKEN_VAULT_URI": "keeper://",
        "OMNITOKEN_CREDENTIALS_NAME": "confluence"
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
mcp-confluence --vault op://MyVault --credentials-name confluence \
               --base-url https://example.atlassian.net/wiki
```

#### Bitwarden Example

```bash
export BW_ACCESS_TOKEN="..."
export BW_ORGANIZATION_ID="..."
mcp-confluence --vault bw://org-id --credentials-name confluence \
               --base-url https://example.atlassian.net/wiki
```

#### Keeper Example

```bash
export KSM_TOKEN="US:..."
mcp-confluence --vault keeper:// --credentials-name confluence \
               --base-url https://example.atlassian.net/wiki
```

### Environment Variables

| Variable | Flag | Description |
|----------|------|-------------|
| `CONFLUENCE_BASE_URL` | `--base-url` | Your Confluence instance URL |
| `CONFLUENCE_USERNAME` | `--username` | Your Confluence username (usually your email) |
| `CONFLUENCE_API_TOKEN` | `--api-token` | API token from [Atlassian Account Settings](https://id.atlassian.com/manage-profile/security/api-tokens) |
| `OMNITOKEN_VAULT_URI` | `--vault` | Vault URI for credentials |
| `OMNITOKEN_CREDENTIALS_NAME` | `--credentials-name` | Name of credentials in vault (default: `confluence`) |
| `OP_SERVICE_ACCOUNT_TOKEN` | - | 1Password service account token |
| `BW_ACCESS_TOKEN` | - | Bitwarden access token |
| `BW_ORGANIZATION_ID` | - | Bitwarden organization ID |
| `KSM_TOKEN` | - | Keeper token (format: `REGION:TOKEN`) |

### Running Standalone (for testing)

```bash
export CONFLUENCE_BASE_URL=https://example.atlassian.net/wiki
export CONFLUENCE_USERNAME=user@example.com
export CONFLUENCE_API_TOKEN=your-api-token

./mcp-confluence
```

## MCP Tools

The MCP server exposes these tools:

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
