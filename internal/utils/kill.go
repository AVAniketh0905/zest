//go:build linux

package utils

import (
	"fmt"
	"syscall"
)

func killWithSyscall(pid int) error {
	err := syscall.Kill(pid, syscall.SIGKILL)
	if err != nil {
		fmt.Printf("Failed to kill PID %d: %v\n", pid, err)
	}
	return err
}

func Kill(pid int) error {
	return killWithSyscall(pid)
}
