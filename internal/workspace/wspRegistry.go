package workspace

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/AVAniketh0905/zest/internal/utils"
)

// WspRegistry maintains a global map of all known workspaces.
// It maps workspace names to their corresponding config file paths.
// This is stored in ~/.zest/state/workspace.json
type WspRegistry struct {
	path string

	Workspaces map[string]string `json:"workspaces"` // key = workspace name, value = path to its config YAML file
}

// NewWspRegistry loads the registry from disk or initializes an empty one.
func NewWspRegistry() (*WspRegistry, error) {
	regPath := filepath.Join(utils.ZestStateDir, "workspaces.json")

	reg := &WspRegistry{
		Workspaces: make(map[string]string),
		path:       regPath,
	}

	// Try to read from disk if exists
	if data, err := os.ReadFile(regPath); err == nil {
		if err := json.Unmarshal(data, &reg.Workspaces); err != nil {
			return nil, err
		}
	}

	return reg, nil
}

// Update adds or updates a workspace in the registry.
func (wr *WspRegistry) Update(cfg *WspConfig) {
	wr.Workspaces[cfg.Name] = cfg.Path
}

// Save writes the current state of the registry to disk.
func (wr *WspRegistry) Save() error {
	if wr.path == "" {
		return errors.New("registry path is not set")
	}
	data, err := json.MarshalIndent(wr, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(wr.path, data, 0644)
}

// GetPath returns the config path for a workspace name.
func (wr *WspRegistry) GetPath(name string) (string, bool) {
	path, ok := wr.Workspaces[name]
	return path, ok
}

// Exists checks if a workspace with the given name exists.
func (wr *WspRegistry) Exists(name string) bool {
	_, ok := wr.Workspaces[name]
	return ok
}
