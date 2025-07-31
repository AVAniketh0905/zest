//go:build windows

package utils

import (
	"errors"
	"log"
	"os/exec"
	"strconv"
)

var (
	errWindowsTaskKill error = errors.New("exit status 128")
)

func killWithTaskkill(pid int) error {
	cmd := exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/T", "/F")
	_, err := cmd.CombinedOutput() // _ -> output

	// fmt.Printf("taskkill output for PID %d:\n%s\n", pid, string(output))
	if errors.Is(err, errWindowsTaskKill) {
		log.Println("process already terminated, ", pid)
		return nil
	}
	return err
}

func Kill(pid int) error {
	return killWithTaskkill(pid)
}
