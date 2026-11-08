package main

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/harsig/harness-kit/gitsource"
	"github.com/harsig/harness-kit/yamlconfig"
)

// Config is the parsed configuration file.
type Config struct {
	Sources []Source `yaml:"sources"`
}

// Source is a local directory or a git repository containing skills.
type Source struct {
	Path    string          `yaml:"path"`
	Git     *gitsource.Repo `yaml:"git"`
	Folders []string        `yaml:"folders"`

	// Root is the resolved absolute source directory, set by LoadConfig.
	Root string `yaml:"-"`
}

// LoadConfig reads, validates and resolves the configuration file at path.
func LoadConfig(path string) (*Config, error) {
	var cfg Config
	configDir, err := yamlconfig.Load(path, &cfg)
	if err != nil {
		return nil, err
	}
	for i := range cfg.Sources {
		if err := resolveRoot(&cfg.Sources[i], configDir); err != nil {
			return nil, fmt.Errorf("%s: sources[%d]: %w", path, i, err)
		}
	}
	return &cfg, nil
}

func resolveRoot(src *Source, configDir string) error {
	switch {
	case src.Path != "" && src.Git != nil:
		return errors.New("only one of path or git is allowed")
	case src.Path != "":
		src.Root = yamlconfig.ResolvePath(configDir, src.Path)
		return nil
	case src.Git != nil:
		if err := src.Git.Validate(); err != nil {
			return err
		}
		src.Git.SetDir(filepath.Join(configDir, ".skills-mcp", "repos"))
		src.Root = src.Git.Dir
		return nil
	default:
		return errors.New("one of path or git is required")
	}
}
