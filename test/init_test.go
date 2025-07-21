package test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/AVAniketh0905/zest/cmd"
)

func setupTempDir() (string, func()) {
	tempDir, _ := os.MkdirTemp("", "zest-test-*")
	return tempDir, func() {
		os.RemoveAll(filepath.Join(tempDir, ".zest"))
	}
}

func TestInitCmd(t *testing.T) {
	tempDir, cleanup := setupTempDir()
	defer cleanup()

	testCases := []struct {
		args     []string
		fail     bool
		expected string
	}{
		{args: []string{"init", "personal"}, fail: true, expected: "Error: required flag(s) \"name\" not set\n"},
		{args: []string{"init"}, fail: true, expected: "Error: required flag(s) \"name\" not set\n"},
		{args: []string{"init", "-n", "work"}, fail: false, expected: "Initialized the workspace, work\n"},
		{args: []string{"init", "--name", "personal"}, fail: false, expected: "Initialized the workspace, personal\n"},
		{args: []string{"init", "-n", "work"}, fail: true, expected: "Error: workspace already exists\n"},
		{args: []string{"init", "-n", ""}, fail: true, expected: "Error: invalid workspace name\n"},
	}

	for _, tc := range testCases {
		rootCmd := cmd.NewRootCmd()
		buf := new(bytes.Buffer)
		errbuf := new(bytes.Buffer)

		tc.args = append(tc.args, "--custom", tempDir)

		rootCmd.SetOut(buf)
		rootCmd.SetErr(errbuf)
		rootCmd.SetArgs(tc.args)
		rootCmd.Execute()

		if tc.fail {
			out := errbuf.String()
			if out != tc.expected {
				t.Log(tc.args)
				t.Logf("buf: %v\n\n, errbuf: %v\n\n", buf.String(), errbuf.String())
				t.Errorf("incorrect output: expected %q, got %q\n", tc.expected, out)
			}
		} else {
			out := buf.String()
			if out != tc.expected {
				t.Log(tc.args)
				t.Logf("buf: %v\n\n, errbuf: %v\n\n", buf.String(), errbuf.String())
				t.Errorf("incorrect output: expected %q, got %q\n", tc.expected, out)
			}
		}
	}
}
