package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"go.yaml.in/yaml/v3"
)

const skillFile = "SKILL.md"

var validName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Skill is a folder containing a SKILL.md.
type Skill struct {
	Name        string
	Description string
	File        string // absolute path to SKILL.md
}

// LoadSkills collects and validates all skills of the configured sources.
func LoadSkills(cfg *Config) ([]Skill, error) {
	var skills []Skill
	seen := map[string]string{}
	for _, src := range cfg.Sources {
		folders := src.Folders
		if len(folders) == 0 {
			folders = []string{"."}
		}
		for _, folder := range folders {
			dir := filepath.Join(src.Root, folder)
			fallback := filepath.Base(dir)
			if src.Git != nil && dir == src.Root {
				fallback = repoName(src.Git.URL) // the clone folder name carries a hash
			}
			s, err := loadSkill(dir, fallback)
			if err != nil {
				return nil, err
			}
			if other, ok := seen[s.Name]; ok {
				return nil, fmt.Errorf("duplicate skill name %q: %s and %s", s.Name, other, s.File)
			}
			seen[s.Name] = s.File
			skills = append(skills, s)
		}
	}
	return skills, nil
}

// loadSkill reads dir/SKILL.md; fallback is the name used when the frontmatter has none.
func loadSkill(dir, fallback string) (Skill, error) {
	file := filepath.Join(dir, skillFile)
	data, err := os.ReadFile(file)
	if err != nil {
		return Skill{}, fmt.Errorf("skill %s: %w", dir, err)
	}
	fm, err := parseFrontmatter(string(data))
	if err != nil {
		return Skill{}, fmt.Errorf("%s: %w", file, err)
	}
	s := Skill{Name: fm.Name, Description: fm.Description, File: file}
	if s.Name == "" {
		s.Name = fallback
	}
	if !validName.MatchString(s.Name) {
		return Skill{}, fmt.Errorf("%s: invalid skill name %q (allowed: %s)", file, s.Name, validName)
	}
	return s, nil
}

// repoName is the repository name of a git URL, e.g. "skills" for git@github.com:acme/skills.git.
func repoName(url string) string {
	return strings.TrimSuffix(path.Base(strings.TrimRight(url, "/")), ".git")
}

type frontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// parseFrontmatter reads name and description from a leading "---" delimited
// YAML block. Content without frontmatter yields empty values.
func parseFrontmatter(content string) (frontmatter, error) {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	if strings.TrimSpace(lines[0]) != "---" {
		return frontmatter{}, nil
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			var fm frontmatter
			if err := yaml.Unmarshal([]byte(strings.Join(lines[1:i], "\n")), &fm); err != nil {
				return frontmatter{}, fmt.Errorf("invalid frontmatter: %w", err)
			}
			return fm, nil
		}
	}
	return frontmatter{}, fmt.Errorf("unterminated frontmatter")
}
