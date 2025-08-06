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

	"github.com/AVAniketh0905/zest/internal/utils"
	"github.com/AVAniketh0905/zest/internal/workspace"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
func NewDeleteCmd(cfg *utils.ZestConfig) *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete [workspace-name]",
		Short: "Delete data for a workspace if it's not running",
		Long: `Deletes the configuration and data for a specific workspace (WSP)
only if it is not currently running. You can use --force to delete even if it's active.`,
		Example: `  zest delete my-workspace
  zest delete my-workspace --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate args early
			if err := cmd.ValidateArgs(args); err != nil {
				return err
			}

			// Check for --force flag
			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				return err
			}

			wspName := args[0]

			// Attempt to delete the workspace
			return deleteWorkspace(cfg, cmd.OutOrStdout(), wspName, force)
		},
	}

	// Optional --force flag
	deleteCmd.Flags().BoolP("force", "f", false, "Force deletion even if the workspace is active")

	return deleteCmd
}

func deleteWorkspace(cfg *utils.ZestConfig, w io.Writer, wspName string, force bool) error {
	// Load workspace registry
	wspReg, err := workspace.NewWspRegistry(cfg)
	if err != nil {
		return fmt.Errorf("unable to load workspace registry: %w", err)
	}

	// Fetch workspace config
	wspCfg, ok := wspReg.GetCfg(wspName)
	if !ok {
		return fmt.Errorf("workspace '%s' not found", wspName)
	}

	// If the workspace is active and force is not enabled, block deletion
	if wspCfg.Status == workspace.Active && !force {
		fmt.Fprintf(w, "Workspace '%s' is active. Use --force to delete it.\n", wspName)
		return workspace.ErrWorkspaceIsActive
	}

	// If force is enabled, attempt to gracefully close the workspace first
	if force && wspCfg.Status == workspace.Active {
		fmt.Fprintf(w, "Force-deleting active workspace '%s'...\n", wspName)
		if err := closeWorkspace(cfg, wspReg, wspCfg); err != nil {
			return fmt.Errorf("failed to close active workspace '%s': %w", wspName, err)
		}
	}

	// Proceed with deletion
	fmt.Fprintf(w, "Deleting workspace '%s'...\n", wspName)
	if err := wspReg.Delete(wspName); err != nil {
		return fmt.Errorf("failed to delete workspace '%s': %w", wspName, err)
	}

	// Save the updated registry
	if err := wspReg.Save(); err != nil {
		return fmt.Errorf("failed to save updated registry: %w", err)
	}

	fmt.Fprintf(w, "Workspace '%s' deleted successfully.\n", wspName)
	return nil
}
