package test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/AVAniketh0905/zest/cmd"
	"github.com/AVAniketh0905/zest/internal/utils"
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
	var currentSection string

outerloop:
	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch line {
		case "INACTIVE WORKSPACES":
			currentSection = "Inactive"
			continue
		case "ACTIVE WORKSPACES":
			currentSection = "Active"
			continue
		case "Extra:":
			break outerloop
		}

		// Skip header
		if strings.HasPrefix(line, "NAME") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue // ignore summary/extras
		}

		row := WorkspaceStatusRow{
			Name:      fields[0],
			Status:    currentSection,
			Timestamp: fields[2],
			Info:      fields[3],
		}

		rows = append(rows, row)
	}
	return rows, nil
}

func TestStatusCommand_ShowsInactiveWorkspaces(t *testing.T) {
	tempDir := setupTempDir(t)

	initCmds := [][]string{
		{"init", "inactive1"},
		{"init", "inactive2"},
	}

	cfg := &utils.ZestConfig{}
	for _, args := range initCmds {
		cmd := cmd.NewRootCmd(cfg)
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		args = append(args, "--custom", tempDir)
		cmd.SetArgs(args)
		require.NoError(t, cmd.Execute())
	}

	cmd := cmd.NewRootCmd(cfg)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"status", "--custom", tempDir})
	require.NoError(t, cmd.Execute())

	rows, err := parseStatusOutput(buf.Bytes())
	require.NoError(t, err)

	names := map[string]bool{}
	for _, row := range rows {
		if row.Status == "Inactive" {
			names[row.Name] = true
			require.NotEmpty(t, row.Info)
		}
	}

	require.True(t, names["inactive1"])
	require.True(t, names["inactive2"])
}

func TestStatusCommand_AllInactive_DefaultBehavior(t *testing.T) {
	tempDir := setupTempDir(t)

	cfg := &utils.ZestConfig{}
	for _, wsp := range []string{"work", "personal"} {
		cmd := cmd.NewRootCmd(cfg)
		cmd.SetArgs([]string{"init", wsp, "--custom", tempDir})
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		require.NoError(t, cmd.Execute())
	}

	cmd := cmd.NewRootCmd(cfg)
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
	tempDir := setupTempDir(t)

	cfg := &utils.ZestConfig{}
	cmd := cmd.NewRootCmd(cfg)
	cmd.SetArgs([]string{"init", "work", "--custom", tempDir})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	require.NoError(t, cmd.Execute())

	var buf bytes.Buffer
	cmd.SetArgs([]string{"status", "home", "nothome", "--custom", tempDir})
	cmd.SetOut(&buf)
	cmd.SetErr(io.Discard)
	require.NoError(t, cmd.Execute())
	output := buf.String()

	require.Contains(t, output, "Skipped: home, nothome\n")
	require.Contains(t, output, "No active workspaces.\n")
	require.NotContains(t, output, "INACTIVE WORKSPACES")
	require.Contains(t, output, "work")
}

func TestStatusCommand_SomeSkippedSomeValid(t *testing.T) {
	tempDir := setupTempDir(t)

	cfg := &utils.ZestConfig{}
	cmd := cmd.NewRootCmd(cfg)
	cmd.SetArgs([]string{"init", "personal", "--custom", tempDir})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	require.NoError(t, cmd.Execute())

	var buf bytes.Buffer
	cmd.SetArgs([]string{"status", "personal", "home", "homenot", "--custom", tempDir})
	cmd.SetOut(&buf)
	cmd.SetErr(io.Discard)
	require.NoError(t, cmd.Execute())
	output := buf.String()

	require.Contains(t, output, "Skipped: home, homenot\n")
	require.Contains(t, output, "personal")
	require.Contains(t, output, "INACTIVE WORKSPACES")
}

func TestStatusCommand_VerboseMode(t *testing.T) {
	tempDir := setupTempDir(t)
	cfg := &utils.ZestConfig{}

	// Initialize and launch a workspace
	cmd := cmd.NewRootCmd(cfg)
	cmd.SetArgs([]string{"init", "verbose-test", "--custom", tempDir})
	require.NoError(t, cmd.Execute())

	cmd.SetArgs([]string{"launch", "verbose-test", "--custom", tempDir})
	require.NoError(t, cmd.Execute())

	var buf bytes.Buffer
	cmd.SetArgs([]string{"status", "--verbose", "--custom", tempDir})
	cmd.SetOut(&buf)
	cmd.SetErr(io.Discard)
	require.NoError(t, cmd.Execute())
	output := buf.String()

	require.Contains(t, output, "ACTIVE WORKSPACES")
	require.Contains(t, output, "Extra Details:")
	require.Contains(t, output, "verbose-test")
	require.Contains(t, output, "Detached:")
}

func TestStatusCommand_JSONMode(t *testing.T) {
	tempDir := setupTempDir(t)
	cfg := &utils.ZestConfig{}

	cmd := cmd.NewRootCmd(cfg)
	cmd.SetArgs([]string{"init", "json-test", "--custom", tempDir})
	require.NoError(t, cmd.Execute())

	var buf bytes.Buffer
	cmd.SetArgs([]string{"status", "--json", "--custom", tempDir})
	cmd.SetOut(&buf)
	cmd.SetErr(io.Discard)
	require.NoError(t, cmd.Execute())
	output := buf.String()

	require.True(t, strings.HasPrefix(output, "{"), "Expected JSON object output")
	require.Contains(t, output, "json-test")
}

// TODO: test watch for statusCmd
