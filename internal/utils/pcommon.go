package utils

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/process"
)

func diff(newPIDs, oldPIDs []int) []int {
	oldMap := make(map[int]bool)
	for _, pid := range oldPIDs {
		oldMap[pid] = true
	}

	var diff []int
	for _, pid := range newPIDs {
		if !oldMap[pid] {
			diff = append(diff, pid)
		}
	}
	return diff
}

func ListPIDs(pname string) ([]int, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("failed to list processes: %w", err)
	}

	var pids []int
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}
		name = strings.ToLower(name)
		if runtime.GOOS == "windows" {
			name = strings.TrimSuffix(name, ".exe")
		}
		if name == pname {
			pids = append(pids, int(p.Pid))
		}
	}
	return pids, nil
}

func WaitForNewPIDs(name string, before []int, maxWait time.Duration) ([]int, error) {
	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		after, err := ListPIDs(name)
		if err != nil {
			return nil, err
		}
		diffPids := diff(after, before)
		if len(diffPids) > 0 {
			return diffPids, nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return nil, errors.New("timed out waiting for new processes")
}
