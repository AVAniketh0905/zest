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
	tempDir, cleanup := setupTempDir()
	defer cleanup()

	// Init workspace
	cmd := cmd.NewRootCmd()
	cmd.SetArgs([]string{"init", "-n", "dev", "--custom", tempDir})
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
	tempDir, cleanup := setupTempDir()
	defer cleanup()

	cmd := cmd.NewRootCmd()
	cmd.SetArgs([]string{"launch", "ghost", "--custom", tempDir})
	cmd.SetOut(io.Discard)

	var bufErr bytes.Buffer
	cmd.SetErr(&bufErr)

	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "workspace does not exist")
}

func TestLaunchCommand_RejectsAlreadyActiveWorkspace(t *testing.T) {
	tempDir, cleanup := setupTempDir()
	defer cleanup()

	// Init
	cmd := cmd.NewRootCmd()
	cmd.SetArgs([]string{"init", "-n", "dev", "--custom", tempDir})
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
	tempDir, cleanup := setupTempDir()
	defer cleanup()

	// Init workspace
	cmd := cmd.NewRootCmd()
	cmd.SetArgs([]string{"init", "-n", "dev", "--custom", tempDir})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	require.NoError(t, cmd.Execute())

	// Launch workspace
	cmd.SetArgs([]string{"launch", "dev", "--custom", tempDir})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	require.NoError(t, cmd.Execute())

	// Check runtime file exists
	rtFile := filepath.Join(utils.ZestRuntimeWspDir(), "dev.json")
	_, err := os.Stat(rtFile)
	require.NoError(t, err, "expected runtime file to be created for launched workspace")

	// Check workspace.json is updated
	wsStateFile := filepath.Join(utils.ZestStateDir(), "workspaces.json")
	data, err := os.ReadFile(wsStateFile)
	require.NoError(t, err)
	require.Contains(t, string(data), `"dev"`)
	require.Contains(t, string(data), `"status": "active"`)
}
