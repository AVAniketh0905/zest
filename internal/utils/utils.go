package utils

import (
	"log"
	"os"
	"path/filepath"
)

type ZestErr error

var (
	ZestWspDir   = filepath.Join(ZestDir(), "workspaces")
	ZestStateDir = filepath.Join(ZestDir(), "state")
)

func ZestDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".zest")
}

func EnsureZestDirs() error {
	dirs := []string{
		ZestWspDir,
		ZestStateDir,
		filepath.Join(ZestStateDir, "workspaces"),
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
