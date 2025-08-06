package launch

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/AVAniketh0905/zest/internal/utils"
)

type BraveApp struct {
	Tabs       []string `yaml:"tabs"`        // List of URLs to open
	ProfileDir string   `yaml:"profile_dir"` // Optional --user-data-dir
	Args       []string `yaml:"args"`        // Optional extra args

	pids []int
	env  map[string]string
}

func (b *BraveApp) GetName() string              { return "brave" }
func (b *BraveApp) GetPIDs() []int               { return b.pids }
func (b *BraveApp) SetEnv(env map[string]string) { b.env = env }

func (b *BraveApp) Start() error {
	bef, err := utils.ListPIDs(b.GetName())
	if err != nil {
		return err
	}

	baseArgs := []string{}
	if b.ProfileDir != "" {
		baseArgs = append(baseArgs, "--user-data-dir="+b.ProfileDir)
	}
	baseArgs = append(baseArgs, b.Args...)
	baseArgs = append(baseArgs, b.Tabs...)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		args := append([]string{"/C", "start", "", "brave.exe"}, baseArgs...)
		cmd = exec.Command("cmd", args...)
	default:
		cmd = exec.Command("brave", baseArgs...)
	}

	if b.env != nil {
		envList := []string{}
		for k, v := range b.env {
			envList = append(envList, k+"="+v)
		}
		cmd.Env = append(cmd.Env, envList...)
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	newPIDs, err := utils.WaitForNewPIDs(b.GetName(), bef, 3*time.Second)
	if err != nil {
		return fmt.Errorf("[zest] warning: couldn't detect new PowerShell process in time")
	}

	b.pids = append(b.pids, newPIDs...)
	return nil
}

func (b *BraveApp) Summary() string {
	out := "- [brave] Process: " + b.GetName() + "\n"
	if len(b.Tabs) > 0 {
		out += "  Tabs:\n"
		for _, tab := range b.Tabs {
			out += "    - " + tab + "\n"
		}
	}
	if b.ProfileDir != "" {
		out += "  Profile Dir: " + b.ProfileDir + "\n"
	}
	if len(b.Args) > 0 {
		out += "  Extra Args: [" + utils.JoinQuoted(b.Args) + "]\n"
	}
	return out
}
