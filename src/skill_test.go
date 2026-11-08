package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/harsig/harness-kit/gitsource"
)

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    frontmatter
		wantErr bool
	}{
		{name: "none", content: "# Title\n", want: frontmatter{}},
		{name: "empty file", content: "", want: frontmatter{}},
		{name: "full", content: "---\nname: pdf\ndescription: Work with PDFs\n---\nbody", want: frontmatter{"pdf", "Work with PDFs"}},
		{name: "crlf", content: "---\r\nname: pdf\r\n---\r\nbody", want: frontmatter{Name: "pdf"}},
		{name: "empty block", content: "---\n---\nbody", want: frontmatter{}},
		{name: "extra keys ignored", content: "---\nname: a\nlicense: MIT\n---\n", want: frontmatter{Name: "a"}},
		{name: "unterminated", content: "---\nname: a\n", wantErr: true},
		{name: "invalid yaml", content: "---\nname: [\n---\n", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFrontmatter(tt.content)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestLoadSkills(t *testing.T) {
	tests := []struct {
		name    string
		files   map[string]string // relative to the temp dir
		sources []Source          // Root is relative to the temp dir
		want    []Skill           // File is relative to the temp dir
		wantErr string
	}{
		{
			name: "folders with frontmatter and fallback",
			files: map[string]string{
				"src/a/SKILL.md": "---\nname: alpha\ndescription: A skill\n---\n",
				"src/b/SKILL.md": "no frontmatter",
			},
			sources: []Source{{Root: "src", Folders: []string{"a", "b"}}},
			want: []Skill{
				{Name: "alpha", Description: "A skill", File: "src/a/SKILL.md"},
				{Name: "b", File: "src/b/SKILL.md"},
			},
		},
		{
			name:    "root is the skill",
			files:   map[string]string{"single/SKILL.md": "x"},
			sources: []Source{{Root: "single"}},
			want:    []Skill{{Name: "single", File: "single/SKILL.md"}},
		},
		{
			name:    "git root without name uses the repository name",
			files:   map[string]string{"one-skill-1a2b3c4d/SKILL.md": "x"},
			sources: []Source{{Root: "one-skill-1a2b3c4d", Git: &gitsource.Repo{URL: "git@github.com:acme/one-skill.git"}}},
			want:    []Skill{{Name: "one-skill", File: "one-skill-1a2b3c4d/SKILL.md"}},
		},
		{
			name:    "missing SKILL.md",
			files:   map[string]string{"src/a/README.md": "x"},
			sources: []Source{{Root: "src", Folders: []string{"a"}}},
			wantErr: "SKILL.md",
		},
		{
			name:    "invalid name",
			files:   map[string]string{"src/SKILL.md": "---\nname: my skill\n---\n"},
			sources: []Source{{Root: "src"}},
			wantErr: "invalid skill name",
		},
		{
			name:    "invalid fallback name",
			files:   map[string]string{"my.skill/SKILL.md": "x"},
			sources: []Source{{Root: "my.skill"}},
			wantErr: "invalid skill name",
		},
		{
			name: "duplicate across sources",
			files: map[string]string{
				"one/a/SKILL.md": "x",
				"two/a/SKILL.md": "x",
			},
			sources: []Source{{Root: "one", Folders: []string{"a"}}, {Root: "two", Folders: []string{"a"}}},
			wantErr: "duplicate skill name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, content := range tt.files {
				writeFile(t, filepath.Join(dir, name), content)
			}
			cfg := &Config{Sources: tt.sources}
			for i := range cfg.Sources {
				cfg.Sources[i].Root = filepath.Join(dir, cfg.Sources[i].Root)
			}

			got, err := LoadSkills(cfg)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d skills, want %d", len(got), len(tt.want))
			}
			for i, want := range tt.want {
				want.File = filepath.Join(dir, want.File)
				if got[i] != want {
					t.Errorf("skill %d = %+v, want %+v", i, got[i], want)
				}
			}
		})
	}
}
