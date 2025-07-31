package launch

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/AVAniketh0905/zest/internal/utils"
	"gopkg.in/yaml.v3"
)

type AppSpec interface {
	GetName() string
	GetPIDs() []int
	Start() error
}

type Plan struct {
	Name       string
	WorkingDir string

	Apps []AppSpec
}

type rawPlanYAML struct {
	Name       string `yaml:"name"`
	WorkingDir string `yaml:"workspace_dir"`

	Apps map[string][]map[string]any `yaml:"apps"` // dynamic decoding
}

func NewLaunchPlan(wspName string) (*Plan, error) {
	path := filepath.Join(utils.ZestWspDir(), wspName+".yaml")

	plan := &Plan{}
	plan.Name = wspName
	if data, err := os.ReadFile(path); err == nil {
		if err := plan.parse(data); err != nil {
			return nil, err
		}
	}

	return plan, nil
}

func (ls *Plan) parse(data []byte) error {
	raw := rawPlanYAML{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return err
	}

	ls.Name = raw.Name
	ls.WorkingDir = raw.WorkingDir
	ls.Apps = []AppSpec{}

	for appType, appList := range raw.Apps {
		for _, appData := range appList {
			appBytes, err := json.Marshal(appData)
			if err != nil {
				return err
			}

			var app AppSpec

			switch appType {
			case "custom":
				var custom CustomApp
				if err := json.Unmarshal(appBytes, &custom); err != nil {
					return err
				}
				app = &custom
			}

			ls.Apps = append(ls.Apps, app)
		}
	}

	return nil
}

// TODO: launches goroutines to start executing apps
func (ls *Plan) Start() error {
	for _, app := range ls.Apps {
		if err := app.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (ls *Plan) GetProcessNames() []string {
	names := []string{}
	for _, app := range ls.Apps {
		names = append(names, app.GetName())
	}
	return names
}

func (ls *Plan) GetPIDs() [][]int {
	pids := [][]int{}
	for _, app := range ls.Apps {
		pids = append(pids, app.GetPIDs())
	}
	return pids
}
