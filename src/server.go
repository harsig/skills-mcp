package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// noArguments is the input schema of every skill tool.
var noArguments = json.RawMessage(`{"type":"object","properties":{},"additionalProperties":false}`)

// readOnly tells clients that loading a skill changes nothing and needs no approval.
var readOnly = &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: new(false)}

// NewServer creates an MCP server with one tool per skill.
func NewServer(skills []Skill) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "skills-mcp", Version: version}, nil)
	for _, s := range skills {
		server.AddTool(&mcp.Tool{
			Name:        s.Name,
			Description: s.Description,
			InputSchema: noArguments,
			Annotations: readOnly,
		}, readSkill(s.File))
	}
	return server
}

// readSkill returns a handler that reads the SKILL.md on every call.
func readSkill(file string) mcp.ToolHandler {
	return func(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := os.ReadFile(file)
		if err != nil {
			res := &mcp.CallToolResult{}
			res.SetError(fmt.Errorf("reading skill: %w", err))
			return res, nil
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(data)}}}, nil
	}
}
