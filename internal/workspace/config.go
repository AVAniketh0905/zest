package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/AVAniketh0905/zest/internal/utils"
	"gopkg.in/yaml.v3"
)

type Status string

var (
	Active   Status = "active"
	Inactive Status = "inactive"
)

var (
	ErrInvalidWorkspaceName utils.ZestErr = errors.New("invalid workspace name")
	ErrWorkspaceExists      utils.ZestErr = errors.New("workspace already exists")
	ErrWorkspaceNotExists   utils.ZestErr = errors.New("workspace does not exist")
)

// WspConfig defines the user-provided configuration for a workspace.
// This is stored as a YAML file at ~/.zest/workspaces/<name>.yaml
type WspConfig struct {
	Name   string `json:"name" yaml:"name"`     // Name of the workspace
	Path   string `json:"path" yaml:"path"`     // Absolute path to the editable workspace config file
	Status Status `json:"status" yaml:"status"` // Current status of the workspace (e.g., Active, Inactive)

	WorkspaceDir string `json:"workspace_dir" yaml:"workspace_dir"` // Root directory where all commands will be executed

	Template string `json:"template,omitempty" yaml:"template,omitempty"` // Optional template name this workspace is based on

	Created     string `json:"created" yaml:"created"`           // Timestamp of when the config file was created (RFC3339 format)
	LastUpdated string `json:"last_updated" yaml:"last_updated"` // Timestamp of the last modification to the config file
	LastUsed    string `json:"last_used" yaml:"last_used"`       // Timestamp of the most recent launch via `zest launch`
}

// returns true if workspace with the given name already exists
func checkName(name string, force bool) error {
	if strings.TrimSpace(name) == "" {
		return ErrInvalidWorkspaceName
	}

	// allows alphanumeric names for workspaces
	reg := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)
	if !reg.MatchString(name) {
		return ErrInvalidWorkspaceName
	}

	file := fmt.Sprintf("%v.yaml", name)
	wspPath := filepath.Join(utils.ZestWspDir(), file)

	if _, err := os.Stat(wspPath); errors.Is(err, os.ErrNotExist) || force {
		return nil
	}

	return ErrWorkspaceExists
}

func Init(name, template string, force bool) error {
	if err := checkName(name, force); err != nil {
		return err
	}

	reg, err := NewWspRegistry()
	if err != nil {
		return fmt.Errorf("failed to create a new registry, %v", err)
	}

	cfg := WspConfig{
		Name:     name,
		Status:   Inactive,
		Template: template,
		Created:  time.Now().Format(time.RFC3339),
		Path:     filepath.Join(utils.ZestWspDir(), name+".yaml"),
		LastUsed: "never",
	}
	cfg.LastUpdated = cfg.Created

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal yaml file, %s", err)
	}

	// write user editable workspace config file
	if err := os.WriteFile(cfg.Path, data, 0644); err != nil {
		return fmt.Errorf("failed to write to worksapce config file at %v, %v", cfg.Path, err)
	}

	reg.Update(&cfg)

	if err := reg.Save(); err != nil {
		return fmt.Errorf("failed to save workspace config, %v", err)
	}

	return nil
}
