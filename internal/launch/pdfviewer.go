package launch

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/AVAniketh0905/zest/internal/utils"
)

type SioyekApp struct {
	Path  string   `yaml:"path"`  // Override default path to binary
	Files []string `yaml:"files"` // Each file will open in a new sioyek instance
	Args  []string `yaml:"args"`

	pids []int
	env  map[string]string
}

func (s *SioyekApp) GetName() string              { return "sioyek" }
func (s *SioyekApp) GetPIDs() []int               { return s.pids }
func (s *SioyekApp) SetEnv(env map[string]string) { s.env = env }

func (s *SioyekApp) Start() error {
	name := s.GetName()

	before, err := utils.ListPIDs(name)
	if err != nil {
		return err
	}

	for _, filePath := range s.Files {
		binaryPath := s.Path
		if binaryPath == "" {
			switch runtime.GOOS {
			case "windows":
				binaryPath = `C:\Program Files\sioyek\sioyek.exe`
			case "darwin":
				binaryPath = "/Applications/sioyek.app/Contents/MacOS/sioyek"
			default: // linux
				binaryPath = "/usr/bin/sioyek"
			}
		}

		args := []string{"--new-window", filePath}
		args = append(args, s.Args...)
		cmd := exec.Command(binaryPath, args...)

		if len(s.env) > 0 {
			envList := os.Environ()
			for k, v := range s.env {
				envList = append(envList, fmt.Sprintf("%s=%s", k, v))
			}
			cmd.Env = envList
		}

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start sioyek for %s: %w", filePath, err)
		}
	}

	newPIDs, err := utils.WaitForNewPIDs(name, before, 2*time.Second)
	if err != nil {
		return fmt.Errorf("[zest] warning: failed to detect new sioyek pids: %v", err)
	}
	s.pids = newPIDs

	return nil
}

func (s *SioyekApp) Summary() string {
	out := "- [sioyek] Process: " + s.GetName() + "\n"
	if len(s.Files) > 0 {
		out += "  Files:\n"
		for _, file := range s.Files {
			out += "    - " + file + "\n"
		}
	}
	return out
}
