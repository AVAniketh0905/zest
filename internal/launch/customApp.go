package launch

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/AVAniketh0905/zest/internal/utils"
)

type CustomApp struct {
	Name string   `yaml:"name"`
	Cmd  string   `yaml:"cmd"`
	Args []string `yaml:"args"`

	pids []int
	env  map[string]string
}

func (c *CustomApp) GetName() string { return c.Name }
func (c *CustomApp) GetPIDs() []int  { return c.pids }
func (c *CustomApp) SetEnv(env map[string]string) {
	c.env = env
}
func (c *CustomApp) Start() error {
	bef, err := utils.ListPIDs(c.Name)
	if err != nil {
		return err
	}

	cmd := exec.Command(c.Cmd, c.Args...)

	if c.env != nil {
		envList := []string{}
		for k, v := range c.env {
			envList = append(envList, k+"="+v)
		}
		cmd.Env = append(cmd.Env, envList...)
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	newPIDs, err := utils.WaitForNewPIDs(c.GetName(), bef, 3*time.Second)
	if err != nil {
		return fmt.Errorf("[zest] warning: couldn't detect new PowerShell process in time")
	}

	c.pids = append(c.pids, newPIDs...)
	return nil
}

func (c *CustomApp) Summary() string {
	formatArgs := func(args []string) string {
		return "[" + utils.JoinQuoted(args) + "]"
	}

	out := "- [custom] " + c.Name + "\n"
	out += "  Cmd: " + c.Cmd + "\n"
	if len(c.Args) > 0 {
		out += "  Args: " + formatArgs(c.Args) + "\n"
	}
	if len(c.env) > 0 {
		out += "  Env:\n"
		for k, v := range c.env {
			out += "    " + k + "=" + v + "\n"
		}
	}
	return out
}
