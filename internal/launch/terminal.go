package launch

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/AVAniketh0905/zest/internal/utils"
)

type PowerShellApp struct {
	Tabs []string `yaml:"tabs"` // custom per-tab commands
	Args []string `yaml:"args"` // e.g., -NoExit

	pids []int
	env  map[string]string

	workingDir string // injected from Plan
}

func (p *PowerShellApp) GetName() string              { return "powershell" }
func (p *PowerShellApp) GetPIDs() []int               { return p.pids }
func (p *PowerShellApp) SetEnv(env map[string]string) { p.env = env }
func (p *PowerShellApp) SetWorkingDir(dir string)     { p.workingDir = dir }

func (p *PowerShellApp) Start() error {
	if runtime.GOOS != "windows" {
		return errors.New("PowerShellApp is only supported on Windows")
	}

	bef, err := utils.ListPIDs(p.GetName())
	if err != nil {
		return err
	}

	// Fallback: open a single tab without a specific command
	tabs := p.Tabs
	if len(tabs) == 0 {
		tabs = []string{""} // just open shell
	}

	for _, tabCmd := range tabs {
		args := []string{"-w", "new-tab"}

		if p.workingDir != "" {
			args = append(args, "--startingDirectory", p.workingDir)
		}

		args = append(args, "powershell.exe")
		args = append(args, p.Args...) // e.g., -NoExit

		if tabCmd != "" {
			args = append(args, "-Command", tabCmd)
		}

		cmd := exec.Command("wt", args...)

		if p.env != nil {
			envList := []string{}
			for k, v := range p.env {
				envList = append(envList, fmt.Sprintf("%s=%s", k, v))
			}
			cmd.Env = append(os.Environ(), envList...)
		}

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start PowerShell tab (%s): %w", tabCmd, err)
		}
	}

	newPIDs, err := utils.WaitForNewPIDs(p.GetName(), bef, 3*time.Second)
	if err != nil {
		return fmt.Errorf("[zest] warning: couldn't detect new PowerShell process in time")
	}

	p.pids = append(p.pids, newPIDs...)
	return nil
}

func (p *PowerShellApp) Summary() string {
	out := "- [powershell]\n"
	if len(p.Tabs) > 0 {
		out += "  Tabs:\n"
		for _, tab := range p.Tabs {
			out += "    - " + tab + "\n"
		}
	} else if p.workingDir != "" {
		out += "  Tabs:\n"
		out += "    - cd " + p.workingDir + "\n"
	}
	if len(p.Args) > 0 {
		out += "  Args: [" + utils.JoinQuoted(p.Args) + "]\n"
	}
	return out
}
