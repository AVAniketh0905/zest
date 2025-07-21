package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AVAniketh0905/zest/internal/utils"
	"gopkg.in/yaml.v3"
)

var (
	ErrInvalidWorkspaceName utils.ZestErr = errors.New("invalid workspace name")
	ErrWorkspaceExists      utils.ZestErr = errors.New("workspace already exists")
	ErrWorkspaceNotExists   utils.ZestErr = errors.New("workspace does not exist")
)

// WspConfig represents the user-defined configuration for a single workspace.
// This is stored as a YAML file under ~/.zest/workspace/<name>.yaml
type WspConfig struct {
	Name string `json:"name" yaml:"name"`
	Path string `json:"workspace_dir" yaml:"workspace_dir"`

	Template string `json:"template,omitempty" yaml:"template,omitempty"`
	Created  string `json:"created" yaml:"created"`
}

// returns true if workspace with the given name already exists
func checkName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrInvalidWorkspaceName
	}

	file := fmt.Sprintf("%v.yaml", name)
	wspPath := filepath.Join(utils.ZestWspDir(), file)

	if _, err := os.Stat(wspPath); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return ErrWorkspaceExists
}

func Init(name, template string) error {
	if err := checkName(name); err != nil {
		return err
	}

	reg, err := NewWspRegistry()
	if err != nil {
		return fmt.Errorf("failed to create a new registry, %v", err)
	}

	cfg := WspConfig{
		Name:     name,
		Template: template,
		Created:  time.Now().Format(time.RFC3339),
		Path:     filepath.Join(utils.ZestWspDir(), name+".yaml"),
	}

	// TODO: for now writing only path to yaml file
	data, err := yaml.Marshal(cfg.Path)
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
