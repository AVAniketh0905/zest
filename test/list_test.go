package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/AVAniketh0905/zest/cmd"
	"github.com/AVAniketh0905/zest/internal/utils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

type WorkspaceRow struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	LastUsed string `json:"last_used"`
	Path     string `json:"path"`
}

func parseWorkspaceListTableOutput(output []byte) ([]WorkspaceRow, error) {
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return nil, errors.New("no workspace data found")
	}

	var rows []WorkspaceRow
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			return nil, fmt.Errorf("expected at least 4 columns, got %d: %q", len(fields), line)
		}
		rows = append(rows, WorkspaceRow{
			Name:     fields[0],
			Status:   fields[1],
			LastUsed: fields[2],
			Path:     fields[3],
		})
	}
	return rows, nil
}

func parseWorkspaceListJSONOutput(output []byte) ([]WorkspaceRow, error) {
	var rows []WorkspaceRow
	err := json.Unmarshal(output, &rows)
	return rows, err
}

func setupAndRun(cmd *cobra.Command, tempDir string, args []string) ([]byte, error) {
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(io.Discard)

	args = append(args, "--custom", tempDir)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return buf.Bytes(), err
}

func TestListCommand_TableOutput_ShowsExpectedWorkspaces(t *testing.T) {
	tempDir := setupTempDir(t)

	cfg := &utils.ZestConfig{}
	rootCmd := cmd.NewRootCmd(cfg)

	// Init two workspaces
	for _, name := range []string{"work", "personal"} {
		_, err := setupAndRun(rootCmd, tempDir, []string{"init", name})
		require.NoError(t, err)
	}

	// Now list them
	output, err := setupAndRun(rootCmd, tempDir, []string{"list"})
	require.NoError(t, err)

	rows, err := parseWorkspaceListTableOutput(output)
	require.NoError(t, err)
	require.Len(t, rows, 2)

	names := map[string]bool{}
	for _, row := range rows {
		names[row.Name] = true
		require.NotEmpty(t, row.Path)
		require.NotEmpty(t, row.LastUsed)
		require.Contains(t, []string{"ACTIVE", "INACTIVE"}, row.Status)
	}

	require.True(t, names["work"])
	require.True(t, names["personal"])
}

func TestListCommand_JSONOutput_IsValid(t *testing.T) {
	tempDir := setupTempDir(t)

	cfg := &utils.ZestConfig{}
	rootCmd := cmd.NewRootCmd(cfg)

	_, err := setupAndRun(rootCmd, tempDir, []string{"init", "dev"})
	require.NoError(t, err)

	output, err := setupAndRun(rootCmd, tempDir, []string{"list", "--json"})
	require.NoError(t, err)

	rows, err := parseWorkspaceListJSONOutput(output)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "dev", rows[0].Name)
}

func TestListCommand_FilterActiveOnly(t *testing.T) {
	tempDir := setupTempDir(t)

	cfg := &utils.ZestConfig{}
	rootCmd := cmd.NewRootCmd(cfg)

	_, err := setupAndRun(rootCmd, tempDir, []string{"init", "team"})
	require.NoError(t, err)

	_, err = setupAndRun(rootCmd, tempDir, []string{"init", "notteam"})
	require.NoError(t, err)

	_, err = setupAndRun(rootCmd, tempDir, []string{"launch", "team"})
	require.NoError(t, err)

	output, err := setupAndRun(rootCmd, tempDir, []string{"list", "--filter", "active"})
	require.NoError(t, err)

	if strings.Contains(string(output), "inactive") && strings.Contains(string(output), "active") {
		t.Error("expected only active but found inactive")
	}

	_, err = setupAndRun(rootCmd, tempDir, []string{"close", "team"})
	require.NoError(t, err)
}

func TestListCommand_SortByName(t *testing.T) {
	tempDir := setupTempDir(t)

	cfg := &utils.ZestConfig{}
	rootCmd := cmd.NewRootCmd(cfg)

	for _, name := range []string{"zebra", "alpha", "omega"} {
		_, err := setupAndRun(rootCmd, tempDir, []string{"init", name})
		require.NoError(t, err)
	}

	output, err := setupAndRun(rootCmd, tempDir, []string{"list", "--sort", "name"})
	require.NoError(t, err)

	rows, err := parseWorkspaceListTableOutput(output)
	require.NoError(t, err)
	require.Len(t, rows, 3)

	var names []string
	for _, r := range rows {
		names = append(names, r.Name)
	}
	require.Equal(t, []string{"alpha", "omega", "zebra"}, names)
}

func TestListCommand_InvalidSortKey(t *testing.T) {
	tempDir := setupTempDir(t)

	cfg := &utils.ZestConfig{}
	rootCmd := cmd.NewRootCmd(cfg)

	_, err := setupAndRun(rootCmd, tempDir, []string{"init", "test"})
	require.NoError(t, err)

	_, err = setupAndRun(rootCmd, tempDir, []string{"list", "--sort", "invalid"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid sort key")
}
