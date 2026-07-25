# Quick Start

## Using with Claude Desktop

### 1. Configure Claude Desktop

Add to your Claude Desktop configuration file:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "confluence": {
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

### 2. Restart Claude Desktop

Close and reopen Claude Desktop to load the new configuration.

### 3. Start Using

You can now ask Claude to interact with Confluence:

- "Read the Confluence page with ID 12345"
- "Search for pages about authentication"
- "Create a new page in TEAM space titled 'Meeting Notes'"
- "Update page 12345 with a new table"

## Using the CLI

The server also works as a CLI tool for testing and scripting:

```bash
# Read a page as structured blocks
mcp-atlassian read-page 12345

# Read a page as raw XHTML
mcp-atlassian read-page-xhtml 12345

# Search for pages
mcp-atlassian search-pages "type=page AND text~'authentication'" --limit 10

# Delete a page
mcp-atlassian delete-page 12345

# Pretty print output
mcp-atlassian read-page 12345 --output pretty
```

## Example Workflows

### Read and Understand a Page

```
You: Read the Confluence page 12345

Claude: I'll read that page for you.
[Uses confluence_read_page tool]

The page "API Design Guidelines" contains:
- Heading: REST API Standards
- Paragraph: Our API follows RESTful principles...
- Table with columns: Endpoint, Method, Description
...
```

### Create a New Page

```
You: Create a meeting notes page in the TEAM space

Claude: I'll create a new meeting notes page.
[Uses confluence_create_page tool with structured blocks]

Created page "Meeting Notes 2025-01-15" with ID 67890
URL: https://example.atlassian.net/spaces/TEAM/pages/67890
```

### Update a Table

```
You: Add a new row to the status table on page 12345

Claude: I'll read the current page and add a new row.
[Uses confluence_read_page, then confluence_update_page]

Updated page with new table row:
| Service | Status |
|---------|--------|
| API     | OK     |
| Auth    | OK     |  <- New row added
```

## Why Structured Blocks?

This server uses structured JSON blocks instead of raw XHTML because:

1. **LLMs produce JSON reliably** - No risk of malformed XHTML
2. **Tables are safe** - Proper `<tbody>` structure is guaranteed
3. **Macros work correctly** - Status badges, code blocks render properly
4. **Round-trip safe** - Read, modify, write without corruption
