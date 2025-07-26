package launch

import (
	"os/exec"
)

type AppSpec interface {
	GetName() string
	GetPID() int
	Start() error
}

type CustomApp struct {
	pid int

	Name string   `yaml:"name"`
	Cmd  string   `yaml:"cmd"`
	Args []string `yaml:"args"`
}

func (c *CustomApp) GetName() string { return c.Name }
func (c *CustomApp) GetPID() int     { return c.pid }
func (c *CustomApp) Start() error {
	// TODO: Actually start the custom app — e.g., spawn a process
	cmd := exec.Command(c.Cmd, c.Args...)
	if err := cmd.Start(); err != nil {
		return err
	}
	c.pid = cmd.Process.Pid
	return nil
}
