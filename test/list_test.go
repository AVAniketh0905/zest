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

type WorkspaceRow struct {
	Name     string
	Status   string
	LastUsed string
	Path     string
}

func parseWorkspaceListOutput(output []byte) ([]WorkspaceRow, error) {
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return nil, errors.New("no workspace data found")
	}

	var rows []WorkspaceRow
	for _, line := range lines[1:] { // skip header
		fields := strings.Fields(line)
		if len(fields) != 4 {
			return nil, fmt.Errorf("expected 4 columns, got %d: %q", len(fields), line)
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

func TestListCommand_ShowsExpectedWorkspaces(t *testing.T) {
	tempDir, cleanup := setupTempDir()
	defer cleanup()

	testCases := []struct {
		args []string
	}{
		{args: []string{"init", "work"}},
		{args: []string{"init", "personal"}},
		{args: []string{"list"}},
	}

	var lastOut []byte

	for _, tc := range testCases {
		rootCmd := cmd.NewRootCmd()
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(io.Discard)
		tc.args = append(tc.args, "--custom", tempDir)
		rootCmd.SetArgs(tc.args)

		err := rootCmd.Execute()
		require.NoError(t, err)
		lastOut = buf.Bytes()
	}

	// only parse list output
	rows, err := parseWorkspaceListOutput(lastOut)
	require.NoError(t, err)
	require.Len(t, rows, 2)

	names := map[string]bool{}
	for _, row := range rows {
		names[row.Name] = true
		require.NotEmpty(t, row.Path)
	}

	require.True(t, names["work"])
	require.True(t, names["personal"])
}
