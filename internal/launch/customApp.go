package launch

import (
	"os/exec"

	"github.com/AVAniketh0905/zest/internal/utils"
)

type CustomApp struct {
	pids []int

	Name string   `yaml:"name"`
	Cmd  string   `yaml:"cmd"`
	Args []string `yaml:"args"`
}

func (c *CustomApp) GetName() string { return c.Name }
func (c *CustomApp) GetPIDs() []int  { return c.pids }
func (c *CustomApp) Start() error {
	bef, err := utils.ListPIDs(c.Name)
	if err != nil {
		return err
	}

	cmd := exec.Command(c.Cmd, c.Args...)
	if err := cmd.Start(); err != nil {
		return err
	}

	c.pids = bef // all process with its name
	return nil
}
