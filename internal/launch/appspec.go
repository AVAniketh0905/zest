package launch

type AppSpec interface {
	GetName() string
	GetPID() int
	Start() error
}

type CustomApp struct {
	pid int

	Name string   `yaml:"name"`
	Args []string `yaml:"args"`
}

func (c *CustomApp) GetName() string { return c.Name }
func (c *CustomApp) GetPID() int     { return c.pid }
func (c *CustomApp) Start() error {
	// TODO: Actually start the custom app — e.g., spawn a process
	c.pid = 9999 // Stubbed PID for now
	return nil
}
