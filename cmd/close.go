/*
Copyright © 2025 AVAniketh0905

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"time"

	"github.com/AVAniketh0905/zest/internal/utils"
	"github.com/AVAniketh0905/zest/internal/workspace"
	"github.com/spf13/cobra"
)

// closeCmd represents the close command
func NewCloseCmd(cfg *utils.ZestConfig) *cobra.Command {
	var closeCmd = &cobra.Command{
		Use:   "close [workspace-name]",
		Short: "Close an existing or active workspace",
		Long:  `Close a specific workspace by name, or use --all to close all workspaces.`,
		Example: `  zest close work
  zest close --all                
  zest close personal`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.ValidateArgs(args); err != nil {
				return err
			}

			all, err := cmd.Flags().GetBool("all")
			if err != nil {
				return err
			}

			wspReg, err := workspace.NewWspRegistry(cfg)
			if err != nil {
				// log.Printf("[zest] error loading workspace registry: %v", err)
				return err
			}
			// log.Printf("[zest] loaded workspace registry")

			if all {
				return closeAllWorkspaces(cfg, wspReg)
			}

			wspCfg, ok := wspReg.GetCfg(args[0])
			if !ok {
				return fmt.Errorf("failed to find workspace with %v", args[0])
			}

			return closeWorkspace(cfg, wspReg, wspCfg)
		},
	}

	closeCmd.Flags().Bool("all", false, "Close all currently open workspaces")

	return closeCmd
}

func closeWorkspace(cfg *utils.ZestConfig, wspReg *workspace.WspRegistry, wspCfg *workspace.WspConfig) error {
	if wspCfg.Status == workspace.Inactive {
		// log.Printf("[zest] workspace '%s' is inactive", wspName)
		return workspace.ErrWorkspaceIsInactive
	}

	wspRt, err := workspace.NewWspRuntime(cfg, wspCfg.Name)
	if err != nil {
		// log.Printf("[zest] failed to initialize workspace runtime: %v", err)
		return err
	}
	if err := wspRt.Load(); err != nil {
		// log.Printf("[zest] failed to load workspace runtime: %v", err)
		return err
	}
	// log.Printf("[zest] initialized workspace runtime")

	// kill all processes
	for _, pids := range wspRt.PIDs {
		for _, newPid := range pids {
			if err := utils.Kill(newPid); err != nil {
				return err
			}
		}
	}

	if err := wspRt.Delete(); err != nil {
		// log.Printf("[zest] failed to delete process, %v", err)
		return err
	}

	wspCfg.Status = workspace.Inactive
	wspCfg.LastUsed = time.Now().Format(time.RFC3339)
	wspReg.Update(wspCfg)
	if err := wspReg.Save(); err != nil {
		// log.Printf("[zest] failed to update workspace status in registry: %v", err)
		return err
	}

	return nil
}

func closeAllWorkspaces(cfg *utils.ZestConfig, wspReg *workspace.WspRegistry) error {
	for _, wspName := range wspReg.GetNames() {
		wspCfg, ok := wspReg.GetCfg(wspName)
		if !ok {
			return fmt.Errorf("failed to find workspace with %v", wspName)
		}

		if wspCfg.Status == workspace.Active {
			if err := closeWorkspace(cfg, wspReg, wspCfg); err != nil {
				return err
			}
		}
	}
	return nil
}
