package workspace

// WspRuntime captures the live state of a running workspace session.
// This is stored as a JSON file at ~/.zest/state/workspaces/<workspace-id>.json
type WspRuntime struct {
	Name      string `json:"name"`       // Name of the workspace (duplicated for quick access)
	StartedAt string `json:"started_at"` // Timestamp when the workspace was launched (RFC3339 format)
	AppCount  int    `json:"app_count"`  // Total number of applications launched during this session

	PIDs      []int    `json:"pids"`      // List of process IDs associated with the workspace
	Processes []string `json:"processes"` // Commands or app names launched as part of this workspace

	Ports       []int    `json:"ports,omitempty"`        // Ports opened by services within the workspace
	BrowserURLs []string `json:"browser_urls,omitempty"` // Web URLs opened by this workspace (if any)

	IsDetached bool `json:"is_detached"` // Indicates if the workspace was launched in detached/background mode
}

// TODO: initialize wsp runtime
func NewWspRuntime() (*WspRuntime, error) {
	return &WspRuntime{}, nil
}
