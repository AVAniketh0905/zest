package utils

import (
	"os"
	"path/filepath"
)

type ZestErr error

var (
	CustomZestDir string
)

func ZestDir() string {
	if CustomZestDir != "" {
		return filepath.Join(CustomZestDir, ".zest")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".zest")
}

func ZestWspDir() string {
	return filepath.Join(ZestDir(), "workspaces")
}

func ZestStateDir() string {
	return filepath.Join(ZestDir(), "state")
}

func EnsureZestDirs() error {
	dirs := []string{
		ZestDir(),
		ZestWspDir(),
		ZestStateDir(),
		filepath.Join(ZestStateDir(), "workspaces"),
	}

	for _, dir := range dirs {
		// if exists skip
		if file, err := os.Stat(dir); err == nil && file.IsDir() {
			// log.Println("dir already exists, ", file.Name())
			continue
		}
		if err := os.Mkdir(dir, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}
