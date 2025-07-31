package launch

import (
	"log"
	"os/exec"
)

type CustomApp struct {
	pid int

	Name string   `yaml:"name"`
	Cmd  string   `yaml:"cmd"`
	Args []string `yaml:"args"`
}

func (c *CustomApp) GetName() string { return c.Name }
func (c *CustomApp) GetPID() int     { return c.pid }
func (c *CustomApp) Start() error {
	log.Println("cmd and stuff: ", c.Cmd, c.Args)
	cmd := exec.Command(c.Cmd, c.Args...)
	if err := cmd.Start(); err != nil {
		return err
	}
	c.pid = cmd.Process.Pid
	return nil
}
