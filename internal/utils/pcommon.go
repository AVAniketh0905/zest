package utils

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/process"
)

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

func Diff(newPIDs, oldPIDs []int) []int {
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
