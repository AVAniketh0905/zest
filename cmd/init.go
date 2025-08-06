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

	"github.com/AVAniketh0905/zest/internal/utils"
	"github.com/AVAniketh0905/zest/internal/workspace"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
func NewInitCmd(cfg *utils.ZestConfig) *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init [workspace-name]",
		Short: "Initialize a new workspace",
		Long: `Initializes a new workspace with the given name, setting up directories
and optionally applying a template for scaffolding.

Workspaces are isolated environments used for organizing different contexts like
work, personal, or learning projects.

Use --force to overwrite an existing workspace.`,
		Example: `  zest init work
  zest init work --template dev-template
  zest init work --force
  zest init personal`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			wspName := args[0]

			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				return err
			}
			template, err := cmd.Flags().GetString("template")
			if err != nil {
				return err
			}

			// Display user-facing info
			fmt.Fprintf(cmd.OutOrStdout(), "Initializing workspace '%s'...\n", wspName)

			if force {
				fmt.Fprintln(cmd.OutOrStdout(), "Force enabled: existing workspace will be overwritten if present.")
			}
			if template != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Using template: %s\n", template)
			}

			if err := workspace.Init(cfg, wspName, template, force); err != nil {
				return fmt.Errorf("failed to initialize workspace '%s': %w", wspName, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Workspace '%s' initialized successfully.\n", wspName)
			return nil
		},
	}

	// Optional Flags
	initCmd.Flags().StringP("template", "t", "", "Template to use for workspace scaffolding")
	initCmd.Flags().BoolP("force", "f", false, "Force initialization even if workspace already exists")

	return initCmd
}
