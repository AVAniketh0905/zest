package test

import (
	"bytes"
	"testing"

	"github.com/AVAniketh0905/zest/cmd"
)

func TestInitName(t *testing.T) {
	testCases := []struct {
		args     []string
		fail     bool
		expected string
	}{
		{args: []string{"init", "personal"}, fail: true, expected: "Error: required flag(s) \"name\" not set\n"},
		{args: []string{"init"}, fail: true, expected: "Error: required flag(s) \"name\" not set\n"},
		{args: []string{"init", "-n", "work"}, fail: false, expected: "Initialized the workspace, work\n"},
		{args: []string{"init", "--name", "personal"}, fail: false, expected: "Initialized the workspace, personal\n"},
	}

	for _, tc := range testCases {
		rootCmd := cmd.NewRootCmd()
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

// TODO: test for checking created wsp
func TestInitWorkspace(t *testing.T) {

}
