package test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/AVAniketh0905/zest/cmd"
	"github.com/stretchr/testify/require"
)

type WorkspaceStatusRow struct {
	Name      string
	Status    string
	Timestamp string // LastUsed (inactive) or StartedAt (active)
	Info      string // Path (inactive) or PIDs (active)
}

func parseStatusOutput(output []byte) ([]WorkspaceStatusRow, error) {
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 3 {
		return nil, errors.New("status output too short")
	}

	var rows []WorkspaceStatusRow
	var inActiveSection bool

	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch line {
		case "INACTIVE":
			inActiveSection = true
			continue
		case "ACTIVE":
			inActiveSection = false
			continue
		}

		// Skip section headers
		if strings.HasPrefix(line, "NAME") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			return nil, fmt.Errorf("expected at least 4 columns, got %d: %q", len(fields), line)
		}

		row := WorkspaceStatusRow{
			Name:      fields[0],
			Status:    fields[1],
			Timestamp: fields[2],
			Info:      fields[3],
		}

		if inActiveSection && row.Status != "Inactive" {
			return nil, fmt.Errorf("expected Inactive status, got %q", row.Status)
		}

		rows = append(rows, row)
	}
	return rows, nil
}

func TestStatusCommand_ShowsInactiveWorkspaces(t *testing.T) {
	tempDir, cleanup := setupTempDir()
	defer cleanup()

	// Set up two inactive workspaces
	initCmds := [][]string{
		{"init", "-n", "inactive1"},
		{"init", "-n", "inactive2"},
	}

	for _, args := range initCmds {
		cmd := cmd.NewRootCmd()
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		args = append(args, "--custom", tempDir)
		cmd.SetArgs(args)
		require.NoError(t, cmd.Execute())
	}

	// Run the status command
	cmd := cmd.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"status", "--custom", tempDir})

	err := cmd.Execute()
	require.NoError(t, err)

	rows, err := parseStatusOutput(buf.Bytes())
	require.NoError(t, err)

	// Verify both inactive workspaces are present
	names := map[string]bool{}
	for _, row := range rows {
		if row.Status == "Inactive" {
			names[row.Name] = true
			require.NotEmpty(t, row.Info) // Path should be present
		}
	}

	require.True(t, names["inactive1"], "inactive1 workspace not found")
	require.True(t, names["inactive2"], "inactive2 workspace not found")
}

func TestStatusCommand_AllInactive_DefaultBehavior(t *testing.T) {
	tempDir, cleanup := setupTempDir()
	defer cleanup()

	// Setup: create two workspaces
	for _, wsp := range []string{"work", "personal"} {
		cmd := cmd.NewRootCmd()
		cmd.SetArgs([]string{"init", "-n", wsp, "--custom", tempDir})
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		require.NoError(t, cmd.Execute())
	}

	// Run `status` without args
	cmd := cmd.NewRootCmd()
	var buf bytes.Buffer
	cmd.SetArgs([]string{"status", "--custom", tempDir})
	cmd.SetOut(&buf)
	cmd.SetErr(io.Discard)

	require.NoError(t, cmd.Execute())

	rows, err := parseStatusOutput(buf.Bytes())
	require.NoError(t, err)

	names := map[string]bool{}
	for _, row := range rows {
		if row.Status == "Inactive" {
			names[row.Name] = true
		}
	}

	require.True(t, names["work"])
	require.True(t, names["personal"])
}

func TestStatusCommand_SkippedInvalidWorkspaces(t *testing.T) {
	tempDir, cleanup := setupTempDir()
	defer cleanup()

	// Create actual workspace so we can detect skipped
	cmd := cmd.NewRootCmd()
	cmd.SetArgs([]string{"init", "-n", "work", "--custom", tempDir})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	require.NoError(t, cmd.Execute())

	var buf bytes.Buffer
	cmd.SetArgs([]string{"status", "home", "nothome", "--custom", tempDir})
	cmd.SetOut(&buf)
	cmd.SetErr(io.Discard)

	require.NoError(t, cmd.Execute())
	output := buf.String()

	require.Contains(t, output, "Skipped *", "should skip all provided names")
	require.Contains(t, output, "INACTIVE")
	require.Contains(t, output, "work")
}

func TestStatusCommand_SomeSkippedSomeValid(t *testing.T) {
	tempDir, cleanup := setupTempDir()
	defer cleanup()

	// Create "personal" workspace
	cmd := cmd.NewRootCmd()
	cmd.SetArgs([]string{"init", "-n", "personal", "--custom", tempDir})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	require.NoError(t, cmd.Execute())

	// Run with one valid and two invalid names
	var buf bytes.Buffer
	cmd.SetArgs([]string{"status", "personal", "home", "homenot", "--custom", tempDir})
	cmd.SetOut(&buf)
	cmd.SetErr(io.Discard)

	require.NoError(t, cmd.Execute())
	output := buf.String()

	require.Contains(t, output, "Skipped home homenot")
	require.Contains(t, output, "personal")
	require.Contains(t, output, "INACTIVE")
}
