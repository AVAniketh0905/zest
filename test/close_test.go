package test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AVAniketh0905/zest/cmd"
	"github.com/AVAniketh0905/zest/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestCloseCommand_RejectsInactiveWorkspace(t *testing.T) {
	tempDir := setupTempDir(t)

	cfg := &utils.ZestConfig{}

	// Init workspace (remains inactive)
	rootCmd := cmd.NewRootCmd(cfg)
	rootCmd.SetArgs([]string{"init", "inactive", "--custom", tempDir})
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	require.NoError(t, rootCmd.Execute())

	// Attempt to close it
	rootCmd.SetArgs([]string{"close", "inactive", "--custom", tempDir})
	rootCmd.SetOut(io.Discard)

	var errBuf bytes.Buffer
	rootCmd.SetErr(&errBuf)

	err := rootCmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "workspace is inactive")
}

func TestCloseCommand_ClosesActiveWorkspace(t *testing.T) {
	tempDir := setupTempDir(t)

	cfg := &utils.ZestConfig{}

	// 1. Init workspace
	rootCmd := cmd.NewRootCmd(cfg)
	rootCmd.SetArgs([]string{"init", "alpha", "--custom", tempDir})
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	require.NoError(t, rootCmd.Execute())

	// 2. Launch workspace
	rootCmd.SetArgs([]string{"launch", "alpha", "--custom", tempDir})
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	require.NoError(t, rootCmd.Execute())

	// 3. Close workspace
	rootCmd.SetArgs([]string{"close", "alpha", "--custom", tempDir})
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	require.NoError(t, rootCmd.Execute())

	// 4. Assert runtime file is deleted
	rtFile := filepath.Join(cfg.RuntimeWspDir(), "alpha.json")
	_, err := os.Stat(rtFile)
	require.Error(t, err, "expected runtime file to be deleted")
	require.True(t, os.IsNotExist(err))

	// 5. Assert workspace.json status is now inactive
	wsStateFile := filepath.Join(cfg.StateDir(), "workspaces.json")
	data, err := os.ReadFile(wsStateFile)
	require.NoError(t, err)

	var wspState struct {
		Workspaces map[string]struct {
			Status   string `json:"status"`
			LastUsed string `json:"last_used"`
		} `json:"workspaces"`
	}
	require.NoError(t, json.Unmarshal(data, &wspState))

	wsp := wspState.Workspaces["alpha"]
	require.Equal(t, "inactive", strings.ToLower(wsp.Status))
	require.NotEqual(t, "never", wsp.LastUsed)
}

func TestCloseCommand_ClosesAllActiveWorkspaces(t *testing.T) {
	tempDir := setupTempDir(t)
	cfg := &utils.ZestConfig{}

	// Init root command
	rootCmd := cmd.NewRootCmd(cfg)
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)

	// Create two workspaces: one active, one inactive
	workspaces := []string{"activeWsp", "inactiveWsp"}

	// 1. Init both
	for _, name := range workspaces {
		rootCmd.SetArgs([]string{"init", name, "--custom", tempDir})
		require.NoError(t, rootCmd.Execute())
	}

	// 2. Launch only the active workspace
	rootCmd.SetArgs([]string{"launch", "activeWsp", "--custom", tempDir})
	require.NoError(t, rootCmd.Execute())

	// 3. Close all
	rootCmd.SetArgs([]string{"close", "--all", "--custom", tempDir})
	require.NoError(t, rootCmd.Execute())

	// 4. Check runtime files
	activeFile := filepath.Join(cfg.RuntimeWspDir(), "activeWsp.json")
	inactiveFile := filepath.Join(cfg.RuntimeWspDir(), "inactiveWsp.json")

	_, err := os.Stat(activeFile)
	require.Error(t, err, "expected runtime file to be deleted")
	require.True(t, os.IsNotExist(err))

	_, err = os.Stat(inactiveFile)
	require.Error(t, err, "expected no runtime file for inactive workspace")
	require.True(t, os.IsNotExist(err))

	// 5. Check workspace statuses
	stateFile := filepath.Join(cfg.StateDir(), "workspaces.json")
	data, err := os.ReadFile(stateFile)
	require.NoError(t, err)

	var wspState struct {
		Workspaces map[string]struct {
			Status   string `json:"status"`
			LastUsed string `json:"last_used"`
		} `json:"workspaces"`
	}
	require.NoError(t, json.Unmarshal(data, &wspState))

	require.Equal(t, "inactive", strings.ToLower(wspState.Workspaces["activeWsp"].Status))
	require.NotEqual(t, "never", wspState.Workspaces["activeWsp"].LastUsed)

	require.Equal(t, "inactive", strings.ToLower(wspState.Workspaces["inactiveWsp"].Status))
	require.Equal(t, "never", wspState.Workspaces["inactiveWsp"].LastUsed)
}
