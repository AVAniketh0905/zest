package test

import (
	"bytes"
	"testing"

	"github.com/AVAniketh0905/zest/cmd"
	"github.com/AVAniketh0905/zest/internal/utils"
)

func TestVersion(t *testing.T) {
	testCases := []struct {
		args     []string
		fail     bool
		expected string
	}{
		{args: []string{"init", "-v"}, fail: true, expected: "Error: unknown shorthand flag: 'v' in -v\n"},
		{args: []string{"init", "--version"}, fail: true, expected: "Error: unknown flag: --version\n"},
		{args: []string{"-v"}, fail: false, expected: "zest: Manage multiple workspaces from a unified CLI - version 0.1.0\n"},
		{args: []string{"--version"}, fail: false, expected: "zest: Manage multiple workspaces from a unified CLI - version 0.1.0\n"},
	}

	cfg := &utils.ZestConfig{}
	for _, tc := range testCases {
		rootCmd := cmd.NewRootCmd(cfg)
		buf := new(bytes.Buffer)
		errbuf := new(bytes.Buffer)

		rootCmd.SetOut(buf)
		rootCmd.SetErr(errbuf)
		rootCmd.SetArgs(tc.args)
		rootCmd.Execute()

		if tc.fail {
			out := errbuf.String()
			if out != tc.expected {
				t.Errorf("incorrect output: expected %q, got %q\n", tc.expected, out)
			}
		} else {
			out := buf.String()
			if out != tc.expected {
				t.Errorf("incorrect output: expected %q, got %q\n", tc.expected, out)
			}
		}
	}
}
