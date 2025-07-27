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

func setupTempDir() (string, func()) {
	tempDir, _ := os.MkdirTemp("", "zest-test-*")
	return tempDir, func() {
		_ = os.RemoveAll(filepath.Join(tempDir, ".zest"))
	}
}

func TestInitCommand_CreatesWorkspaceSuccessfully(t *testing.T) {
	tempDir, cleanup := setupTempDir()
	defer cleanup()

	cmd := cmd.NewRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"init", "work", "--custom", tempDir})

	err := cmd.Execute()
	require.NoError(t, err)
	require.Equal(t, "Initialized the workspace, work\n", buf.String())

	// Assert workspace file created
	wspFile := filepath.Join(utils.ZestWspDir(), "work.yaml")
	_, err = os.Stat(wspFile)
	require.NoError(t, err)
}

func TestInitCommand_RejectsMissingNameFlag(t *testing.T) {
	tempDir, cleanup := setupTempDir()
	defer cleanup()

	cmd := cmd.NewRootCmd()
	var errBuf bytes.Buffer
	cmd.SetOut(io.Discard)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"init", "--custom", tempDir})

	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, errBuf.String(), `Error: requires at least 1 arg(s), only received 0`)
}

func TestInitCommand_RejectsEmptyWorkspaceName(t *testing.T) {
	tempDir, cleanup := setupTempDir()
	defer cleanup()

	cmd := cmd.NewRootCmd()
	var errBuf bytes.Buffer
	cmd.SetOut(io.Discard)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"init", "", "--custom", tempDir})

	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, errBuf.String(), "Error: invalid workspace name\n")
}

func TestInitCommand_RejectsDuplicateWorkspace(t *testing.T) {
	tempDir, cleanup := setupTempDir()
	defer cleanup()

	// First creation
	cmd := cmd.NewRootCmd()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"init", "duplicate", "--custom", tempDir})
	require.NoError(t, cmd.Execute())

	// Second creation
	var errBuf bytes.Buffer
	cmd.SetOut(io.Discard)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"init", "duplicate", "--custom", tempDir})

	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, errBuf.String(), "workspace already exists")
}
