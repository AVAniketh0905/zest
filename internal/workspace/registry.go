package workspace

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/AVAniketh0905/zest/internal/utils"
)

// WspRegistry maintains a global map of all known workspaces.
// This is stored in ~/.zest/state/workspace.json
type WspRegistry struct {
	path string

	Workspaces map[string]*WspConfig `json:"workspaces"` // key = workspace name, value = path to its config YAML file
}

// NewWspRegistry loads the registry from disk or initializes an empty one.
func NewWspRegistry(cfg *utils.ZestConfig) (*WspRegistry, error) {
	regPath := filepath.Join(cfg.StateDir(), "workspaces.json")

	reg := &WspRegistry{
		path:       regPath,
		Workspaces: map[string]*WspConfig{},
	}

	// Try to read from disk if exists
	if data, err := os.ReadFile(regPath); err == nil {
		if err := json.Unmarshal(data, &reg); err != nil {
			return nil, err
		}
	}

	return reg, nil
}

// Update adds or updates a workspace in the registry.
func (wr *WspRegistry) Update(cfg *WspConfig) {
	wr.Workspaces[cfg.Name] = cfg
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

// GetCfg return a pointer to `WspConfig` struct for the workspace name.
func (wr *WspRegistry) GetCfg(name string) (*WspConfig, bool) {
	cfg, ok := wr.Workspaces[name]
	return cfg, ok
}

// GetPath returns the config path for a workspace name.
func (wr *WspRegistry) GetPath(name string) (string, bool) {
	cfg, ok := wr.Workspaces[name]
	return cfg.Path, ok
}

// GetNames returns slice of names of the workspaces
func (wr *WspRegistry) GetNames() []string {
	out := []string{}
	for _, wsp := range wr.Workspaces {
		out = append(out, wsp.Name)
	}
	return out
}

// Exists checks if a workspace with the given name exists.
func (wr *WspRegistry) Exists(name string) bool {
	_, ok := wr.Workspaces[name]
	return ok
}
