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

var (
	template string
	force    bool
)

// initCmd represents the init command
func NewInitCmd(cfg *utils.ZestConfig) *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init [workspace-name]",
		Short: "Initialize a new workspace",
		Long: `The init command creates a new workspace with the given name, setting up the
necessary directory structure and optional template files.

Workspaces are isolated environments used for organizing different contexts like
work, personal, or learning projects.

You can optionally specify a template to scaffold the workspace with predefined files.
Use --force to overwrite existing workspaces if necessary.
`,
		Example: `  zest init work 
  zest init work --template [template-name]
  zest init work --force
  zest init personal`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.ValidateArgs(args); err != nil {
				return err
			}
			wspName := args[0] // TODO: multiple workspaces

			if force {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Force enabled: existing workspaces will be overwritten.")
			}

			if template != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Using template: %s\n", template)
			}

			err := workspace.Init(cfg, wspName, template, force) // TODO: template
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Initialized the workspace,", wspName)
			return err
		},
	}

	initCmd.Flags().StringVarP(&template, "template", "t", "", "Template to use for workspace scaffolding")
	initCmd.Flags().BoolVarP(&force, "force", "f", false, "Force initialization even if workspace already exists")

	return initCmd
}
