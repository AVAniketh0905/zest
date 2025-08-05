package launch

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/AVAniketh0905/zest/internal/utils"
)

type VSCodeApp struct {
	Path string   `yaml:"path"`           // project folder to open
	Args []string `yaml:"args,omitempty"` // additional args

	pids       []int
	env        map[string]string
	workingDir string // set by plan
}

func (v *VSCodeApp) GetName() string              { return "code" }
func (v *VSCodeApp) GetPIDs() []int               { return v.pids }
func (v *VSCodeApp) SetEnv(env map[string]string) { v.env = env }
func (v *VSCodeApp) SetWorkingDir(dir string)     { v.workingDir = dir }

func (v *VSCodeApp) Start() error {
	bef, err := utils.ListPIDs(v.GetName())
	if err != nil {
		return err
	}

	args := []string{}
	if v.Path != "" {
		args = append(args, v.Path)
	} else if v.Path == "" && v.workingDir != "" {
		args = append(args, v.workingDir)
	}

	if len(v.Args) > 0 {
		args = append(args, v.Args...)
	}

	// Assuming VSCode is in PATH
	cmd := exec.Command("code", args...)

	if v.env != nil {
		envList := []string{}
		for k, v := range v.env {
			envList = append(envList, fmt.Sprintf("%s=%s", k, v))
		}
		cmd.Env = append(cmd.Env, envList...)
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	newPIDs, err := utils.WaitForNewPIDs(v.GetName(), bef, 3*time.Second)
	if err != nil {
		return fmt.Errorf("[zest] warning: couldn't detect new vscode process in time")
	}

	v.pids = append(v.pids, newPIDs...)
	return nil
}

func (v *VSCodeApp) Summary() string {
	out := "- [vscode] Opening folder: " + v.Path + "\n"
	if len(v.Args) > 0 {
		out += "  Args: [" + utils.JoinQuoted(v.Args) + "]\n"
	}
	return out
}
