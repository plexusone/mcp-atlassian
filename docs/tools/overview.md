# Tools Reference

This server provides 26 tools for working with Confluence pages and Jira issues.

## Confluence Tools

### Reading Pages

### confluence_read_page

Read a page as structured blocks.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `page_id` | string | Yes | The page ID |

**Example:**

```json
{
  "name": "confluence_read_page",
  "arguments": {
    "page_id": "12345"
  }
}
```

**Response:**

```json
{
  "page_id": "12345",
  "title": "API Design Guidelines",
  "version": 5,
  "blocks": [
    {"type": "heading", "level": 1, "text": "Overview"},
    {"type": "paragraph", "text": "This document describes..."},
    {"type": "table", "headers": ["Method", "Path"], "rows": [...]}
  ]
}
```

### confluence_read_page_xhtml

Read a page as raw Storage Format XHTML.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `page_id` | string | Yes | The page ID |

**Response:**

```json
{
  "page_id": "12345",
  "title": "API Design Guidelines",
  "version": 5,
  "xhtml": "<h1>Overview</h1><p>This document describes...</p>"
}
```

Use this when you need to preserve complex formatting or debug parsing issues.

### Writing Pages

### confluence_create_page

Create a new page with structured blocks.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `space_key` | string | Yes | Space key (e.g., "TEAM") |
| `title` | string | Yes | Page title |
| `parent_id` | string | No | Parent page ID |
| `blocks` | array | Yes | Content blocks |

**Example:**

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
      {"type": "bullet_list", "items": ["Review PR #123", "Update documentation"]}
    ]
  }
}
```

### confluence_update_page

Update a page with structured blocks.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `page_id` | string | Yes | The page ID |
| `title` | string | No | New title (optional) |
| `blocks` | array | Yes | Content blocks |

**Example:**

```json
{
  "name": "confluence_update_page",
  "arguments": {
    "page_id": "12345",
    "title": "Updated Page Title",
    "blocks": [
      {"type": "heading", "level": 1, "text": "Updated Content"},
      {"type": "paragraph", "text": "This page has been updated."}
    ]
  }
}
```

### confluence_update_page_xhtml

Update a page with raw Storage Format XHTML.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `page_id` | string | Yes | The page ID |
| `title` | string | No | New title (optional) |
| `xhtml` | string | Yes | Raw XHTML content |

Use this when you need to preserve complex formatting.

### Tables

### confluence_create_table

Create a table block from structured data.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `headers` | array | Yes | Column headers |
| `rows` | array | Yes | Row data (2D array) |

**Example:**

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

### Other Operations

### confluence_delete_page

Delete a page.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `page_id` | string | Yes | The page ID |

### confluence_search_pages

Search pages using CQL (Confluence Query Language).

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `cql` | string | Yes | CQL query |
| `limit` | integer | No | Maximum results |

**Example:**

```json
{
  "name": "confluence_search_pages",
  "arguments": {
    "cql": "space=TEAM and type=page and title~\"Meeting\"",
    "limit": 10
  }
}
```

## Jira Tools

### Issues

| Tool | Description |
|------|-------------|
| `jira_get_issue` | Get a Jira issue by key |
| `jira_search` | Search issues using JQL |
| `jira_create_issue` | Create a new issue |
| `jira_update_issue` | Update issue fields |
| `jira_clone_issue` | Clone an existing issue |
| `jira_bulk_update` | Bulk update multiple issues |
| `jira_add_comment` | Add a comment to an issue |
| `jira_get_comments` | Get comments on an issue |
| `jira_get_transitions` | List available transitions for an issue |
| `jira_transition_issue` | Transition an issue to a new status |
| `jira_move_to_sprint` | Move issues to a sprint |

### Projects & Boards

| Tool | Description |
|------|-------------|
| `jira_get_projects` | List available projects |
| `jira_get_boards` | List Scrum/Kanban boards |
| `jira_get_sprints` | List sprints for a board |

### Reports

| Tool | Description |
|------|-------------|
| `jira_velocity_report` | Velocity report for a board |
| `jira_burndown_report` | Sprint burndown report |
| `jira_worklog_report` | Work log report for issues |
| `jira_cycle_time_report` | Cycle time analysis |
