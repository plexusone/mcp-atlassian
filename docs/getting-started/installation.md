# Installation

## Requirements

- Go 1.24 or later
- An Atlassian Cloud or Data Center instance (Confluence, Jira, or both)
- Atlassian API token (Cloud) or credentials (Data Center)

## Install from Source

```bash
go install github.com/plexusone/mcp-atlassian/cmd/mcp-atlassian@latest
```

## Build from Source

```bash
git clone https://github.com/plexusone/mcp-atlassian.git
cd mcp-atlassian
go build ./cmd/mcp-atlassian
```

## Verify Installation

```bash
mcp-atlassian version
```

## As a Library

You can also use the packages directly:

```bash
go get github.com/plexusone/mcp-atlassian
```

### Available Packages

| Package | Description |
|---------|-------------|
| `skills/confluence` | Confluence omniskill — 8 tools for page operations |
| `skills/jira` | Jira omniskill — 18 tools for issue tracking, agile, and reporting |

The underlying Confluence and Jira client libraries live in [Go-Atlassian](https://github.com/grokify/go-atlassian).
