package utils

import (
	"os"
	"path/filepath"
)

type ZestErr error

type ZestConfig struct {
	ZestDir string // Root path to zest directory (default is $HOME)
}

// Root zest directory
func (cfg *ZestConfig) RootDir() string {
	if cfg.ZestDir != "" {
		return filepath.Join(cfg.ZestDir, ".zest")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".zest")
}

// Directory containing all workspaces
func (cfg *ZestConfig) WspDir() string {
	return filepath.Join(cfg.RootDir(), "workspaces")
}

// Directory storing state
func (cfg *ZestConfig) StateDir() string {
	return filepath.Join(cfg.RootDir(), "state")
}

// Directory storing runtime info of workspaces
func (cfg *ZestConfig) RuntimeWspDir() string {
	return filepath.Join(cfg.StateDir(), "workspaces")
}

// Ensures all necessary directories exist
func (cfg *ZestConfig) EnsureDirs() error {
	dirs := []string{
		cfg.RootDir(),
		cfg.WspDir(),
		cfg.StateDir(),
		cfg.RuntimeWspDir(),
	}

	for _, dir := range dirs {
		if file, err := os.Stat(dir); err == nil && file.IsDir() {
			continue
		}
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}
