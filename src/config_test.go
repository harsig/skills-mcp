package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harsig/harness-kit/gitsource"
)

// writeFile creates path (and parent directories) with content.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadConfig(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name      string
		yaml      string
		wantRoots []string // relative roots are joined with the config dir
		wantErr   string
	}{
		{name: "empty", yaml: "", wantRoots: nil},
		{name: "relative path", yaml: "sources:\n  - path: ./a\n", wantRoots: []string{"a"}},
		{name: "absolute path", yaml: "sources:\n  - path: /x/y\n", wantRoots: []string{"/x/y"}},
		{name: "tilde", yaml: "sources:\n  - path: ~/s\n", wantRoots: []string{filepath.Join(home, "s")}},
		{
			name:      "git",
			yaml:      "sources:\n  - git:\n      url: git@github.com:acme/skills.git\n      ref: v1\n    folders: [a, b]\n",
			wantRoots: []string{repoDir("git@github.com:acme/skills.git", "v1")},
		},
		{name: "unknown top-level key", yaml: "foo: 1\n", wantErr: "field foo not found"},
		{name: "unknown source key", yaml: "sources:\n  - path: a\n    folder: [x]\n", wantErr: "field folder not found"},
		{name: "unknown git key", yaml: "sources:\n  - git:\n      url: u/x\n      branch: main\n", wantErr: "field branch not found"},
		{name: "path and git", yaml: "sources:\n  - path: a\n    git:\n      url: u/x\n", wantErr: "only one of path or git"},
		{name: "neither", yaml: "sources:\n  - folders: [a]\n", wantErr: "one of path or git is required"},
		{name: "git without url", yaml: "sources:\n  - git:\n      ref: main\n", wantErr: "git.url is required"},
		{name: "invalid yaml", yaml: "sources: [", wantErr: "yaml"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "config.yaml")
			writeFile(t, path, tt.yaml)

			cfg, err := LoadConfig(path)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(cfg.Sources) != len(tt.wantRoots) {
				t.Fatalf("got %d sources, want %d", len(cfg.Sources), len(tt.wantRoots))
			}
			for i, want := range tt.wantRoots {
				if !filepath.IsAbs(want) {
					want = filepath.Join(dir, want)
				}
				if got := cfg.Sources[i].Root; got != want {
					t.Errorf("source %d root = %q, want %q", i, got, want)
				}
			}
		})
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	if _, err := LoadConfig(filepath.Join(t.TempDir(), "nope.yaml")); err == nil {
		t.Fatal("expected error")
	}
}

// repoDir is the clone location of a git source, relative to the config directory.
func repoDir(url, ref string) string {
	r := gitsource.Repo{URL: url, Ref: ref}
	r.SetDir(filepath.Join(".skills-mcp", "repos"))
	return r.Dir
}
