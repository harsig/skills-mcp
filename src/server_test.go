package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// connect starts a server for skills and returns a connected client session.
func connect(t *testing.T, skills []Skill) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	serverT, clientT := mcp.NewInMemoryTransports()
	if _, err := NewServer(skills).Connect(ctx, serverT, nil); err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test"}, nil)
	session, err := client.Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { session.Close() })
	return session
}

func TestServerListTools(t *testing.T) {
	session := connect(t, []Skill{
		{Name: "alpha", Description: "A skill", File: "unused"},
		{Name: "beta", File: "unused"},
	})
	res, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, tool := range res.Tools {
		got[tool.Name] = tool.Description
	}
	want := map[string]string{"alpha": "A skill", "beta": ""}
	if len(got) != len(want) {
		t.Fatalf("tools = %v, want %v", got, want)
	}
	for name, desc := range want {
		if d, ok := got[name]; !ok || d != desc {
			t.Errorf("tool %q: description %q (present %v), want %q", name, d, ok, desc)
		}
	}
	for _, tool := range res.Tools {
		schema, _ := tool.InputSchema.(map[string]any)
		if schema["additionalProperties"] != false || schema["properties"] == nil {
			t.Errorf("tool %q: schema = %v, want an object without properties", tool.Name, tool.InputSchema)
		}
		if a := tool.Annotations; a == nil || !a.ReadOnlyHint || !a.IdempotentHint {
			t.Errorf("tool %q: annotations = %+v, want read-only and idempotent", tool.Name, a)
		}
	}
}

func TestServerCallTool(t *testing.T) {
	file := filepath.Join(t.TempDir(), skillFile)
	session := connect(t, []Skill{{Name: "alpha", File: file}})

	tests := []struct {
		name      string
		content   *string // nil removes the file
		wantText  string
		wantError bool
	}{
		{name: "initial content", content: ptr("---\nname: alpha\n---\nv1"), wantText: "---\nname: alpha\n---\nv1"},
		{name: "edit is visible without restart", content: ptr("v2"), wantText: "v2"},
		{name: "vanished file is a tool error", content: nil, wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.content != nil {
				writeFile(t, file, *tt.content)
			} else if err := os.Remove(file); err != nil {
				t.Fatal(err)
			}

			res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "alpha"})
			if err != nil {
				t.Fatal(err)
			}
			if res.IsError != tt.wantError {
				t.Fatalf("IsError = %v, want %v", res.IsError, tt.wantError)
			}
			if tt.wantError {
				return
			}
			if len(res.Content) != 1 {
				t.Fatalf("got %d content blocks, want 1", len(res.Content))
			}
			text, ok := res.Content[0].(*mcp.TextContent)
			if !ok || text.Text != tt.wantText {
				t.Errorf("content = %#v, want text %q", res.Content[0], tt.wantText)
			}
		})
	}
}

func ptr(s string) *string { return &s }
