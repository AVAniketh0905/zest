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
	"io"
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
		Long: `Close a specific workspace by name, or use --all to close all workspaces.

Closing a workspace will stop its processes and mark it as inactive.`,
		Example: `  zest close work
  zest close personal
  zest close --all`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.ValidateArgs(args); err != nil {
				return err
			}

			// Parse --all flag
			all, err := cmd.Flags().GetBool("all")
			if err != nil {
				return err
			}

			// Load the workspace registry
			wspReg, err := workspace.NewWspRegistry(cfg)
			if err != nil {
				return fmt.Errorf("unable to load workspace registry: %w", err)
			}

			if all {
				return closeAllWorkspaces(cfg, wspReg, cmd.OutOrStdout())
			}

			// Handle specific workspace close
			if len(args) == 0 {
				return fmt.Errorf("workspace name is required unless --all is used")
			}

			wspName := args[0]
			wspCfg, ok := wspReg.GetCfg(wspName)
			if !ok {
				return fmt.Errorf("workspace '%s' not found", wspName)
			}

			if err := closeWorkspace(cfg, wspReg, wspCfg, cmd.OutOrStdout()); err != nil {
				return fmt.Errorf("failed to close workspace '%s': %w", wspName, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Workspace '%s' closed successfully.\n", wspName)
			return nil
		},
	}

	closeCmd.Flags().Bool("all", false, "Close all currently open workspaces")
	return closeCmd
}

func closeWorkspace(cfg *utils.ZestConfig, wspReg *workspace.WspRegistry, wspCfg *workspace.WspConfig, w io.Writer) error {
	// Skip if already inactive
	if wspCfg.Status == workspace.Inactive {
		fmt.Fprintf(w, "Workspace '%s' is already inactive.\n", wspCfg.Name)
		return workspace.ErrWorkspaceIsInactive
	}

	// Initialize runtime for the workspace
	wspRt, err := workspace.NewWspRuntime(cfg, wspCfg.Name)
	if err != nil {
		return fmt.Errorf("failed to initialize runtime for '%s': %w", wspCfg.Name, err)
	}

	if err := wspRt.Load(); err != nil {
		return fmt.Errorf("failed to load runtime for '%s': %w", wspCfg.Name, err)
	}

	// Kill all associated processes
	for _, pids := range wspRt.PIDs {
		for _, pid := range pids {
			if err := utils.Kill(pid); err != nil {
				fmt.Fprintf(w, "Warning: failed to kill PID %d for workspace '%s': %v\n", pid, wspCfg.Name, err)
			}
		}
	}

	// Cleanup any leftover state or files
	if err := wspRt.Delete(); err != nil {
		return fmt.Errorf("failed to cleanup runtime for '%s': %w", wspCfg.Name, err)
	}

	// Update registry
	wspCfg.Status = workspace.Inactive
	wspCfg.LastUsed = time.Now().Format(time.RFC3339)
	wspReg.Update(wspCfg)

	if err := wspReg.Save(); err != nil {
		return fmt.Errorf("failed to update registry: %w", err)
	}

	return nil
}

func closeAllWorkspaces(cfg *utils.ZestConfig, wspReg *workspace.WspRegistry, w io.Writer) error {
	anyClosed := false

	for _, wspName := range wspReg.GetNames() {
		wspCfg, ok := wspReg.GetCfg(wspName)
		if !ok {
			fmt.Fprintf(w, "Skipping unknown workspace '%s'\n", wspName)
			continue
		}

		if wspCfg.Status == workspace.Active {
			fmt.Fprintf(w, "Closing workspace '%s'...\n", wspName)
			if err := closeWorkspace(cfg, wspReg, wspCfg, w); err != nil {
				fmt.Fprintf(w, "Error closing workspace '%s': %v\n", wspName, err)
				continue
			}
			fmt.Fprintf(w, "Workspace '%s' closed successfully.\n", wspName)
			anyClosed = true
		} else {
			fmt.Fprintf(w, "Skipping '%s' — already inactive.\n", wspName)
		}
	}

	if !anyClosed {
		fmt.Fprintln(w, "No active workspaces found to close.")
	}

	return nil
}
