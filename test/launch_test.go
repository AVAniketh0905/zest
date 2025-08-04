package test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/AVAniketh0905/zest/cmd"
	"github.com/AVAniketh0905/zest/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestLaunchCommand_LaunchesInactiveWorkspace(t *testing.T) {
	tempDir := setupTempDir(t)

	// Init workspace
	cfg := &utils.ZestConfig{}
	cmd := cmd.NewRootCmd(cfg)
	cmd.SetArgs([]string{"init", "dev", "--custom", tempDir})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	require.NoError(t, cmd.Execute())

	// Launch it
	cmd.SetArgs([]string{"launch", "dev", "--custom", tempDir})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	require.NoError(t, cmd.Execute())

	// Verify it's now active
	var buf bytes.Buffer
	cmd.SetArgs([]string{"status", "--custom", tempDir})
	cmd.SetOut(&buf)
	cmd.SetErr(io.Discard)
	require.NoError(t, cmd.Execute())

	output := buf.String()
	require.Contains(t, output, "ACTIVE")
	require.Contains(t, output, "dev")
}

func TestLaunchCommand_FailsForNonExistentWorkspace(t *testing.T) {
	tempDir := setupTempDir(t)

	cfg := &utils.ZestConfig{}
	cmd := cmd.NewRootCmd(cfg)
	cmd.SetArgs([]string{"launch", "ghost", "--custom", tempDir})
	cmd.SetOut(io.Discard)

	var bufErr bytes.Buffer
	cmd.SetErr(&bufErr)

	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "workspace does not exist")
}

func TestLaunchCommand_RejectsAlreadyActiveWorkspace(t *testing.T) {
	tempDir := setupTempDir(t)

	// Init
	cfg := &utils.ZestConfig{}
	cmd := cmd.NewRootCmd(cfg)
	cmd.SetArgs([]string{"init", "dev", "--custom", tempDir})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	require.NoError(t, cmd.Execute())

	// First launch
	cmd.SetArgs([]string{"launch", "dev", "--custom", tempDir})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	require.NoError(t, cmd.Execute())

	// Second launch should fail
	cmd.SetArgs([]string{"launch", "dev", "--custom", tempDir})
	cmd.SetOut(io.Discard)

	var errBuf bytes.Buffer
	cmd.SetErr(&errBuf)

	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "workspace already active")
}

func TestLaunchCommand_CreatesRuntimeFile(t *testing.T) {
	tempDir := setupTempDir(t)

	// Init workspace
	cfg := &utils.ZestConfig{}
	cmd := cmd.NewRootCmd(cfg)
	cmd.SetArgs([]string{"init", "dev", "--custom", tempDir})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	require.NoError(t, cmd.Execute())

	// Launch workspace
	cmd.SetArgs([]string{"launch", "dev", "--custom", tempDir})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	require.NoError(t, cmd.Execute())

	// Check runtime file exists
	rtFile := filepath.Join(cfg.RuntimeWspDir(), "dev.json")
	_, err := os.Stat(rtFile)
	require.NoError(t, err, "expected runtime file to be created for launched workspace")

	// Check workspace.json is updated
	wsStateFile := filepath.Join(cfg.StateDir(), "workspaces.json")
	data, err := os.ReadFile(wsStateFile)
	require.NoError(t, err)
	require.Contains(t, string(data), `"dev"`)
	require.Contains(t, string(data), `"status": "active"`)
}

func TestLaunchCommand_RespectsDryRunFlag(t *testing.T) {
	tempDir := setupTempDir(t)

	cfg := &utils.ZestConfig{}
	cmd := cmd.NewRootCmd(cfg)

	// Init
	cmd.SetArgs([]string{"init", "dev", "--custom", tempDir})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	require.NoError(t, cmd.Execute())

	// Launch with dry-run
	var buf bytes.Buffer
	cmd.SetArgs([]string{"launch", "dev", "--dry-run", "--custom", tempDir})
	cmd.SetOut(&buf)
	cmd.SetErr(io.Discard)
	require.NoError(t, cmd.Execute())

	output := buf.String()
	require.Contains(t, output, "Launch Plan Summary")
	require.Contains(t, output, "Workspace: dev")

	// Runtime file should not exist
	rtFile := filepath.Join(cfg.RuntimeWspDir(), "dev.json")
	_, err := os.Stat(rtFile)
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
}

func TestLaunchCommand_OverridesEnv(t *testing.T) {
	tempDir := setupTempDir(t)

	cfg := &utils.ZestConfig{}
	cmd := cmd.NewRootCmd(cfg)

	// Init
	cmd.SetArgs([]string{"init", "dev", "--custom", tempDir})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	require.NoError(t, cmd.Execute())

	// Write simple app to dev.yaml
	yamlPath := filepath.Join(cfg.WspDir(), "dev.yaml")
	yamlContent := []byte(`
name: dev
apps:
  custom:
    - name: echoapp
      cmd: echo
      args: ["Hello", "World"]
`)
	require.NoError(t, os.WriteFile(yamlPath, yamlContent, 0644))

	// Launch with dry-run and env
	var buf bytes.Buffer
	cmd.SetArgs([]string{"launch", "dev", "--custom", tempDir, "--dry-run", "--env", "MODE=test", "--env", "DEBUG=true"})
	cmd.SetOut(&buf)
	cmd.SetErr(io.Discard)
	require.NoError(t, cmd.Execute())

	out := buf.String()
	require.Contains(t, out, "Env:")
	require.Contains(t, out, "MODE=test")
	require.Contains(t, out, "DEBUG=true")
}

func TestLaunchCommand_ForceLaunchesActiveWorkspace(t *testing.T) {
	tempDir := setupTempDir(t)

	cfg := &utils.ZestConfig{}
	cmd := cmd.NewRootCmd(cfg)

	// Init
	cmd.SetArgs([]string{"init", "dev", "--custom", tempDir})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	require.NoError(t, cmd.Execute())

	// First launch
	cmd.SetArgs([]string{"launch", "dev", "--custom", tempDir})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	require.NoError(t, cmd.Execute())

	// Second launch with --force
	cmd.SetArgs([]string{"launch", "dev", "--custom", tempDir, "--force"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	require.NoError(t, cmd.Execute())
}
