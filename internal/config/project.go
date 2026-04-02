package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ProjectConfig struct {
	Model        string            `json:"model,omitempty"`
	System       string            `json:"system,omitempty"`
	Permissions  map[string]string `json:"permissions,omitempty"`
	ExcludePaths []string          `json:"excludePaths,omitempty"`
	Source       string            `json:"-"` // where the config was loaded from
}

func LoadProjectConfig(dir string) (*ProjectConfig, error) {
	// try COAH.md first
	coahMD := filepath.Join(dir, "COAH.md")
	if data, err := os.ReadFile(coahMD); err == nil {
		content := string(data)
		cfg := &ProjectConfig{
			System: extractSystemContent(content),
			Source: coahMD,
		}
		return cfg, nil
	}

	// try .coah/config.json
	configJSON := filepath.Join(dir, ".coah", "config.json")
	if data, err := os.ReadFile(configJSON); err == nil {
		var cfg ProjectConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", configJSON, err)
		}
		cfg.Source = configJSON
		return &cfg, nil
	}

	return nil, nil
}

// extractSystemContent strips optional YAML frontmatter from COAH.md
// and returns the markdown body as a system prompt addition.
func extractSystemContent(content string) string {
	// check for YAML frontmatter
	if strings.HasPrefix(content, "---\n") {
		end := strings.Index(content[4:], "\n---\n")
		if end >= 0 {
			return strings.TrimSpace(content[4+end+5:])
		}
	}
	return strings.TrimSpace(content)
}
