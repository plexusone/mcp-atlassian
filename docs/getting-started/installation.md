# Installation

## Requirements

- Go 1.24 or later
- A Confluence Cloud or Data Center instance
- Confluence API token (Cloud) or credentials (Data Center)

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
| `storage` | IR types, render, parse, validate for Confluence Storage Format |
| `confluence` | REST API client with IR integration |
| `mcpserver` | MCP server implementation |
