// Package confluence provides an omniskill Skill for reading and writing Confluence pages.
//
// This package can be used standalone with mcp-atlassian or composed
// with other skills in a multi-service MCP server.
package confluence

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grokify/go-atlassian/confluence"
	"github.com/grokify/go-atlassian/confluence/storage"
	"github.com/plexusone/omniskill/skill"
)

// Skill provides Confluence reading and writing tools.
type Skill struct {
	client *confluence.Client
}

// New creates a new Confluence skill with the given client.
func New(client *confluence.Client) *Skill {
	return &Skill{client: client}
}

// Name returns the skill identifier.
func (s *Skill) Name() string {
	return "confluence"
}

// Description returns what this skill does.
func (s *Skill) Description() string {
	return "Read, create, update, and search Confluence pages using structured content blocks"
}

// Init initializes the skill (no-op as client is injected).
func (s *Skill) Init(ctx context.Context) error {
	return nil
}

// Close releases resources (no-op for this skill).
func (s *Skill) Close() error {
	return nil
}

// Tools returns all tools provided by this skill.
func (s *Skill) Tools() []skill.Tool {
	return []skill.Tool{
		s.readPageTool(),
		s.readPageXHTMLTool(),
		s.updatePageTool(),
		s.updatePageXHTMLTool(),
		s.createPageTool(),
		s.createTableTool(),
		s.deletePageTool(),
		s.searchPagesTool(),
	}
}

// Ensure Skill implements skill.Skill.
var _ skill.Skill = (*Skill)(nil)

func (s *Skill) readPageTool() skill.Tool {
	return skill.NewTool(
		"confluence_read_page",
		"Read a Confluence page as structured content blocks. Returns the page content parsed into blocks (paragraphs, tables, headings, etc.) that can be safely modified.",
		map[string]skill.Parameter{
			"page_id": {
				Type:        "string",
				Description: "The Confluence page ID",
				Required:    true,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			pageID, _ := params["page_id"].(string)
			if pageID == "" {
				return nil, fmt.Errorf("page_id is required")
			}

			page, info, err := s.client.GetPageStorage(ctx, pageID)
			if err != nil {
				return nil, err
			}

			return map[string]any{
				"page_id":   info.ID,
				"title":     info.Title,
				"version":   info.Version,
				"space_key": info.SpaceKey,
				"blocks":    blocksToJSON(page.Blocks),
			}, nil
		},
	)
}

func (s *Skill) readPageXHTMLTool() skill.Tool {
	return skill.NewTool(
		"confluence_read_page_xhtml",
		"Read a Confluence page as raw Storage Format XHTML. Returns the unparsed XHTML body for debugging or when the block parser doesn't support certain content.",
		map[string]skill.Parameter{
			"page_id": {
				Type:        "string",
				Description: "The Confluence page ID",
				Required:    true,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			pageID, _ := params["page_id"].(string)
			if pageID == "" {
				return nil, fmt.Errorf("page_id is required")
			}

			xhtml, info, err := s.client.GetPageStorageRaw(ctx, pageID)
			if err != nil {
				return nil, err
			}

			return map[string]any{
				"page_id":   info.ID,
				"title":     info.Title,
				"version":   info.Version,
				"space_key": info.SpaceKey,
				"xhtml":     xhtml,
			}, nil
		},
	)
}

func (s *Skill) updatePageTool() skill.Tool {
	return skill.NewTool(
		"confluence_update_page",
		"Update a Confluence page with structured content blocks. Accepts an array of blocks (paragraphs, tables, headings, etc.) and safely renders them to valid Confluence Storage XHTML.",
		map[string]skill.Parameter{
			"page_id": {
				Type:        "string",
				Description: "The Confluence page ID",
				Required:    true,
			},
			"title": {
				Type:        "string",
				Description: "The page title",
				Required:    true,
			},
			"blocks": {
				Type:        "array",
				Description: "Array of content blocks",
				Required:    true,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			pageID, _ := params["page_id"].(string)
			title, _ := params["title"].(string)
			blocksRaw, _ := params["blocks"].([]any)

			if pageID == "" {
				return nil, fmt.Errorf("page_id is required")
			}
			if title == "" {
				return nil, fmt.Errorf("title is required")
			}

			page, err := parseBlocks(blocksRaw)
			if err != nil {
				return nil, fmt.Errorf("invalid blocks: %w", err)
			}

			// Get current version
			_, info, err := s.client.GetPageStorageRaw(ctx, pageID)
			if err != nil {
				return nil, fmt.Errorf("failed to get current version: %w", err)
			}

			if err := s.client.UpdatePageStorage(ctx, pageID, page, info.Version, title); err != nil {
				return nil, err
			}

			return map[string]any{
				"status":  "updated",
				"page_id": pageID,
				"title":   title,
				"version": info.Version + 1,
			}, nil
		},
	)
}

func (s *Skill) updatePageXHTMLTool() skill.Tool {
	return skill.NewTool(
		"confluence_update_page_xhtml",
		"Update a Confluence page with raw Storage Format XHTML. Use this when you need to preserve all formatting, attributes, and structure that the block-based update would lose.",
		map[string]skill.Parameter{
			"page_id": {
				Type:        "string",
				Description: "The Confluence page ID",
				Required:    true,
			},
			"title": {
				Type:        "string",
				Description: "The page title",
				Required:    true,
			},
			"xhtml": {
				Type:        "string",
				Description: "The raw Storage Format XHTML content",
				Required:    true,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			pageID, _ := params["page_id"].(string)
			title, _ := params["title"].(string)
			xhtml, _ := params["xhtml"].(string)

			if pageID == "" {
				return nil, fmt.Errorf("page_id is required")
			}
			if title == "" {
				return nil, fmt.Errorf("title is required")
			}
			if xhtml == "" {
				return nil, fmt.Errorf("xhtml is required")
			}

			// Get current version
			_, info, err := s.client.GetPageStorageRaw(ctx, pageID)
			if err != nil {
				return nil, fmt.Errorf("failed to get current version: %w", err)
			}

			if err := s.client.UpdatePageStorageRaw(ctx, pageID, xhtml, info.Version, title); err != nil {
				return nil, err
			}

			return map[string]any{
				"status":  "updated",
				"page_id": pageID,
				"title":   title,
				"version": info.Version + 1,
			}, nil
		},
	)
}

func (s *Skill) createPageTool() skill.Tool {
	return skill.NewTool(
		"confluence_create_page",
		"Create a new Confluence page with structured content blocks. Accepts an array of blocks and safely renders them to valid Confluence Storage XHTML.",
		map[string]skill.Parameter{
			"space_key": {
				Type:        "string",
				Description: "The space key where the page will be created",
				Required:    true,
			},
			"title": {
				Type:        "string",
				Description: "The page title",
				Required:    true,
			},
			"blocks": {
				Type:        "array",
				Description: "Array of content blocks",
				Required:    true,
			},
			"parent_id": {
				Type:        "string",
				Description: "Optional parent page ID",
				Required:    false,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			spaceKey, _ := params["space_key"].(string)
			title, _ := params["title"].(string)
			parentID, _ := params["parent_id"].(string)
			blocksRaw, _ := params["blocks"].([]any)

			if spaceKey == "" {
				return nil, fmt.Errorf("space_key is required")
			}
			if title == "" {
				return nil, fmt.Errorf("title is required")
			}

			page, err := parseBlocks(blocksRaw)
			if err != nil {
				return nil, fmt.Errorf("invalid blocks: %w", err)
			}

			pageID, err := s.client.CreatePage(ctx, spaceKey, title, page, parentID)
			if err != nil {
				return nil, err
			}

			return map[string]any{
				"status":    "created",
				"page_id":   pageID,
				"title":     title,
				"space_key": spaceKey,
			}, nil
		},
	)
}

func (s *Skill) createTableTool() skill.Tool {
	return skill.NewTool(
		"confluence_create_table",
		"Create a table block from structured data. Returns a table block that can be included in page content.",
		map[string]skill.Parameter{
			"headers": {
				Type:        "array",
				Description: "Column headers",
				Required:    true,
			},
			"rows": {
				Type:        "array",
				Description: "Table rows, each row is an array of cells",
				Required:    true,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			headersRaw, _ := params["headers"].([]any)
			rowsRaw, _ := params["rows"].([]any)

			table := &storage.Table{
				Headers: make([]string, 0, len(headersRaw)),
				Rows:    make([]storage.Row, 0, len(rowsRaw)),
			}

			for _, h := range headersRaw {
				if str, ok := h.(string); ok {
					table.Headers = append(table.Headers, str)
				}
			}

			for _, rowRaw := range rowsRaw {
				row := storage.Row{Cells: []storage.Cell{}}
				cells, ok := rowRaw.([]any)
				if !ok {
					continue
				}
				for _, cellRaw := range cells {
					cell := parseCell(cellRaw)
					row.Cells = append(row.Cells, cell)
				}
				table.Rows = append(table.Rows, row)
			}

			// Validate the table
			if err := storage.ValidateBlock(table); err != nil {
				return nil, fmt.Errorf("table validation failed: %w", err)
			}

			// Render to XHTML for preview
			xhtml, err := storage.RenderBlock(table)
			if err != nil {
				return nil, err
			}

			return map[string]any{
				"block": map[string]any{
					"type":    "table",
					"headers": table.Headers,
					"rows":    table.Rows,
				},
				"xhtml": xhtml,
			}, nil
		},
	)
}

func (s *Skill) deletePageTool() skill.Tool {
	return skill.NewTool(
		"confluence_delete_page",
		"Delete a Confluence page by ID.",
		map[string]skill.Parameter{
			"page_id": {
				Type:        "string",
				Description: "The Confluence page ID to delete",
				Required:    true,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			pageID, _ := params["page_id"].(string)
			if pageID == "" {
				return nil, fmt.Errorf("page_id is required")
			}

			if err := s.client.DeletePage(ctx, pageID); err != nil {
				return nil, err
			}

			return map[string]any{
				"status":  "deleted",
				"page_id": pageID,
			}, nil
		},
	)
}

func (s *Skill) searchPagesTool() skill.Tool {
	return skill.NewTool(
		"confluence_search_pages",
		"Search for Confluence pages using CQL (Confluence Query Language).",
		map[string]skill.Parameter{
			"cql": {
				Type:        "string",
				Description: "CQL query string (e.g., 'space=TEST and type=page')",
				Required:    true,
			},
			"limit": {
				Type:        "integer",
				Description: "Maximum number of results (default 25)",
				Required:    false,
				Default:     25,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			cql, _ := params["cql"].(string)
			if cql == "" {
				return nil, fmt.Errorf("cql is required")
			}

			limit := 25
			if l, ok := params["limit"].(float64); ok {
				limit = int(l)
			}

			pages, err := s.client.SearchPages(ctx, cql, limit)
			if err != nil {
				return nil, err
			}

			results := make([]map[string]any, len(pages))
			for i, p := range pages {
				results[i] = map[string]any{
					"page_id": p.ID,
					"title":   p.Title,
					"type":    p.Type,
					"status":  p.Status,
				}
			}

			return map[string]any{
				"count":   len(results),
				"results": results,
			}, nil
		},
	)
}

// blocksToJSON converts a slice of storage.Block to JSON-serializable format.
func blocksToJSON(blocks []storage.Block) []any {
	result := make([]any, 0, len(blocks))
	for _, block := range blocks {
		result = append(result, blockToJSON(block))
	}
	return result
}

// blockToJSON converts a single storage.Block to JSON-serializable format.
func blockToJSON(block storage.Block) map[string]any {
	switch b := block.(type) {
	case *storage.Table:
		rows := make([]any, len(b.Rows))
		for i, row := range b.Rows {
			cells := make([]any, len(row.Cells))
			for j, cell := range row.Cells {
				if cell.Macro != nil {
					cells[j] = map[string]any{
						"macro": map[string]any{
							"name":   cell.Macro.Name,
							"params": cell.Macro.Params,
						},
					}
				} else {
					cells[j] = cell.Text
				}
			}
			rows[i] = cells
		}
		return map[string]any{
			"type":    "table",
			"headers": b.Headers,
			"rows":    rows,
		}
	case *storage.Paragraph:
		return map[string]any{
			"type": "paragraph",
			"text": b.Text,
		}
	case *storage.Heading:
		return map[string]any{
			"type":  "heading",
			"level": b.Level,
			"text":  b.Text,
		}
	case *storage.BulletList:
		items := make([]string, len(b.Items))
		for i, item := range b.Items {
			items[i] = item.Text
		}
		return map[string]any{
			"type":  "bullet_list",
			"items": items,
		}
	case *storage.NumberedList:
		items := make([]string, len(b.Items))
		for i, item := range b.Items {
			items[i] = item.Text
		}
		return map[string]any{
			"type":  "numbered_list",
			"items": items,
		}
	case *storage.Macro:
		return map[string]any{
			"type":   "macro",
			"name":   b.Name,
			"params": b.Params,
			"body":   b.Body,
		}
	case *storage.CodeBlock:
		return map[string]any{
			"type":     "code_block",
			"language": b.Language,
			"code":     b.Code,
		}
	case *storage.HorizontalRule:
		return map[string]any{
			"type": "horizontal_rule",
		}
	default:
		return map[string]any{
			"type": "unknown",
		}
	}
}

// parseBlocks converts JSON input to a storage.Page.
func parseBlocks(blocksRaw []any) (*storage.Page, error) {
	page := &storage.Page{Blocks: []storage.Block{}}
	for _, b := range blocksRaw {
		block, err := parseBlock(b)
		if err != nil {
			return nil, err
		}
		if block != nil {
			page.Blocks = append(page.Blocks, block)
		}
	}
	return page, nil
}

// parseBlock converts a single JSON block to a storage.Block.
func parseBlock(raw any) (storage.Block, error) {
	m, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("block must be an object")
	}

	blockType, _ := m["type"].(string)

	switch blockType {
	case "table":
		return parseTableBlock(m), nil
	case "paragraph":
		text, _ := m["text"].(string)
		return &storage.Paragraph{Text: text}, nil
	case "heading":
		level := 1
		if l, ok := m["level"].(float64); ok {
			level = int(l)
		}
		text, _ := m["text"].(string)
		return &storage.Heading{Level: level, Text: text}, nil
	case "bullet_list":
		return parseBulletListBlock(m), nil
	case "numbered_list":
		return parseNumberedListBlock(m), nil
	case "macro":
		return parseMacroBlock(m), nil
	case "code_block":
		language, _ := m["language"].(string)
		code, _ := m["code"].(string)
		return &storage.CodeBlock{Language: language, Code: code}, nil
	case "horizontal_rule":
		return &storage.HorizontalRule{}, nil
	default:
		return nil, fmt.Errorf("unknown block type: %s", blockType)
	}
}

func parseTableBlock(m map[string]any) *storage.Table {
	table := &storage.Table{
		Headers: []string{},
		Rows:    []storage.Row{},
	}

	if headers, ok := m["headers"].([]any); ok {
		for _, h := range headers {
			if s, ok := h.(string); ok {
				table.Headers = append(table.Headers, s)
			}
		}
	}

	if rows, ok := m["rows"].([]any); ok {
		for _, rowRaw := range rows {
			row := storage.Row{Cells: []storage.Cell{}}
			if cells, ok := rowRaw.([]any); ok {
				for _, cellRaw := range cells {
					cell := parseCell(cellRaw)
					row.Cells = append(row.Cells, cell)
				}
			}
			table.Rows = append(table.Rows, row)
		}
	}

	return table
}

func parseBulletListBlock(m map[string]any) *storage.BulletList {
	list := &storage.BulletList{Items: []storage.ListItem{}}
	if items, ok := m["items"].([]any); ok {
		for _, item := range items {
			if s, ok := item.(string); ok {
				list.Items = append(list.Items, storage.ListItem{Text: s})
			}
		}
	}
	return list
}

func parseNumberedListBlock(m map[string]any) *storage.NumberedList {
	list := &storage.NumberedList{Items: []storage.ListItem{}}
	if items, ok := m["items"].([]any); ok {
		for _, item := range items {
			if s, ok := item.(string); ok {
				list.Items = append(list.Items, storage.ListItem{Text: s})
			}
		}
	}
	return list
}

func parseMacroBlock(m map[string]any) *storage.Macro {
	macro := &storage.Macro{
		Params: make(map[string]string),
	}
	macro.Name, _ = m["name"].(string)
	macro.Body, _ = m["body"].(string)

	if params, ok := m["params"].(map[string]any); ok {
		for k, v := range params {
			if s, ok := v.(string); ok {
				macro.Params[k] = s
			}
		}
	}

	return macro
}

func parseCell(raw any) storage.Cell {
	switch v := raw.(type) {
	case string:
		return storage.Cell{Text: v}
	case map[string]any:
		if text, ok := v["text"].(string); ok {
			return storage.Cell{Text: text}
		}
		if macroRaw, ok := v["macro"].(map[string]any); ok {
			macro := &storage.Macro{
				Params: make(map[string]string),
			}
			macro.Name, _ = macroRaw["name"].(string)
			if params, ok := macroRaw["params"].(map[string]any); ok {
				for k, val := range params {
					if s, ok := val.(string); ok {
						macro.Params[k] = s
					}
				}
			}
			return storage.Cell{Macro: macro}
		}
	}
	return storage.Cell{}
}

// MarshalToolResult marshals the tool result as JSON text for MCP response.
func MarshalToolResult(result any) (string, error) {
	text, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(text), nil
}
