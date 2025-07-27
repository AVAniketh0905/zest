package workspace

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/AVAniketh0905/zest/internal/launch"
	"github.com/AVAniketh0905/zest/internal/utils"
)

var (
	ErrWorkspaceIsActive   utils.ZestErr = errors.New("workspace already active")
	ErrWorkspaceIsInactive utils.ZestErr = errors.New("workspace is inactive")
)

// WspRuntime captures the live state of a running workspace session.
// This is stored as a JSON file at ~/.zest/state/workspaces/<workspace-id>.json
type WspRuntime struct {
	sync.Mutex

	Name      string `json:"name"`       // Name of the workspace (duplicated for quick access)
	RtFile    string `json:"-"`          // Runtime filepath
	StartedAt string `json:"started_at"` // Timestamp when the workspace was launched (RFC3339 format)
	AppCount  int    `json:"app_count"`  // Total number of applications launched during this session

	PIDs      []int    `json:"pids"`      // List of process IDs associated with the workspace
	Processes []string `json:"processes"` // Commands or app names launched as part of this workspace

	Ports       []int    `json:"ports,omitempty"`        // Ports opened by services within the workspace
	BrowserURLs []string `json:"browser_urls,omitempty"` // Web URLs opened by this workspace (if any)

	IsDetached bool `json:"is_detached"` // Indicates if the workspace was launched in detached/background mode
}

func NewWspRuntime(wspName string) (*WspRuntime, error) {
	wspRt := &WspRuntime{}
	wspRt.Name = wspName
	wspRt.IsDetached = false // TODO: for now set to false
	wspRt.RtFile = filepath.Join(utils.ZestRuntimeWspDir(), wspName+".json")
	return wspRt, nil
}

func (wspRt *WspRuntime) Load() error {
	// Try to read from disk if exists
	wspRt.Lock()
	defer wspRt.Unlock()

	if data, err := os.ReadFile(wspRt.RtFile); err == nil {
		if err := json.Unmarshal(data, wspRt); err != nil {
			return err
		}
	}
	return nil
}

func (wspRt *WspRuntime) Monitor() error {
	return nil
}

func (wspRt *WspRuntime) Update(plan *launch.Plan) {
	wspRt.StartedAt = time.Now().Format(time.RFC3339)
	wspRt.AppCount = len(plan.Apps)
	wspRt.PIDs = plan.GetPIDs()
	wspRt.Processes = plan.GetProcesNames()
}

func (wspRt *WspRuntime) Save() error {
	if wspRt.RtFile == "" {
		return errors.New("registry path is not set")
	}
	data, err := json.MarshalIndent(wspRt, "", "  ")
	if err != nil {
		return err
	}

	wspRt.Lock()
	defer wspRt.Unlock()

	return os.WriteFile(wspRt.RtFile, data, 0644)
}

func (wspRt *WspRuntime) Delete() error {
	wspRt.Lock()
	defer wspRt.Unlock()

	return os.Remove(wspRt.RtFile)
}
