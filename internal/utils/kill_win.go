//go:build windows

package utils

import (
	"os/exec"
	"strconv"
)

func killWithTaskkill(pid int) error {
	cmd := exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/T", "/F")
	_, err := cmd.CombinedOutput() // _ -> output

	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() == 128 {
			// log.Printf("Taskkill returned exit code 128 (likely: process does not exist), pid: %d", pid)
			return nil // consider it non-fatal
		}
	}

	return err
}

func Kill(pid int) error {
	return killWithTaskkill(pid)
}
