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

	"github.com/AVAniketh0905/zest/internal/launch"
	"github.com/AVAniketh0905/zest/internal/utils"
	"github.com/AVAniketh0905/zest/internal/workspace"
	"github.com/spf13/cobra"
)

type LaunchOptions struct {
	DryRun bool
	Detach bool
	Env    map[string]string
	Force  bool
}

// launchCmd represents the launch command
func NewLaunchCmd(cfg *utils.ZestConfig) *cobra.Command {
	var launchCmd = &cobra.Command{
		Use:   "launch [workspace-name]",
		Short: "Launch a workspace",
		Long: `Launches the specified workspace, initializing its runtime state and executing
its startup command.
`,
		Args: cobra.ExactArgs(1),
		Example: `  zest launch work
  zest launch work --detach
  zest launch personal --dry-run
  zest launch work --env MODE=dev
  zest launch work --force
  zest launch personal --dry-run --env MODE=test`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.ValidateArgs(args); err != nil {
				return err
			}

			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				return err
			}

			detach, err := cmd.Flags().GetBool("detach")
			if err != nil {
				return err
			}

			env, err := cmd.Flags().GetStringToString("env")
			if err != nil {
				return err
			}

			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				return err
			}

			opts := LaunchOptions{
				DryRun: dryRun,
				Detach: detach,
				Env:    env,
				Force:  force,
			}

			return launchWorkspace(cmd.OutOrStdout(), cfg, opts, args[0])
		},
	}

	launchCmd.Flags().Bool("dry-run", false, "Validate config and simulate launch without executing")
	launchCmd.Flags().BoolP("detach", "d", false, "Run workspace in background")
	launchCmd.Flags().StringToString("env", nil, "Set or override environment variables (e.g. --env KEY=VALUE)")
	launchCmd.Flags().BoolP("force", "f", false, "Force launch even if workspace is active")

	return launchCmd
}

func launchWorkspace(w io.Writer, cfg *utils.ZestConfig, opts LaunchOptions, wspName string) error {
	// log.Printf("[zest] starting launch sequence for workspace: %s", wspName)

	wspReg, err := workspace.NewWspRegistry(cfg)
	if err != nil {
		// log.Printf("[zest] error loading workspace registry: %v", err)
		return err
	}
	// log.Printf("[zest] loaded workspace registry")

	wspCfg, ok := wspReg.GetCfg(wspName)
	if !ok {
		// log.Printf("[zest] workspace '%s' does not exist", wspName)
		return workspace.ErrWorkspaceNotExists
	}

	if wspCfg.Status == workspace.Active && !opts.Force {
		// log.Printf("[zest] workspace '%s' is already active", wspName)
		return workspace.ErrWorkspaceIsActive
	}
	// log.Printf("[zest] workspace '%s' is valid and inactive", wspName)

	plan, err := launch.NewLaunchPlan(cfg, wspName)
	if err != nil {
		// log.Printf("[zest] failed to parse launch plan: %v", err)
		return err
	}
	// log.Printf("[zest] parsed launch plan for workspace '%s'", wspName)

	if len(opts.Env) > 0 {
		plan.ApplyEnv(opts.Env)
	}

	wspRt, err := workspace.NewWspRuntime(cfg, wspCfg.Name)
	if err != nil {
		// log.Printf("[zest] failed to initialize workspace runtime: %v", err)
		return err
	}
	// log.Printf("[zest] initialized workspace runtime")

	if opts.DryRun {
		fmt.Printf("[zest] Dry-run mode: skipping actual launch. Launch plan:\n\n")
		fmt.Fprintln(w, plan.Summary())
		return nil
	}

	if err := plan.Start(); err != nil {
		// log.Printf("[zest] failed to launch apps: %v", err)
		return err
	}
	// log.Printf("[zest] launched apps successfully")

	wspRt.Update(plan)
	// log.Printf("[zest] updated workspace runtime state")

	if err := wspRt.Save(); err != nil {
		// log.Printf("[zest] failed to save workspace runtime state: %v", err)
		return err
	}
	// log.Printf("[zest] saved workspace runtime state")

	wspCfg.Status = workspace.Active
	wspReg.Update(wspCfg)
	if err := wspReg.Save(); err != nil {
		// log.Printf("[zest] failed to update workspace status in registry: %v", err)
		return err
	}
	// log.Printf("[zest] workspace '%s' marked as active and registry saved", wspName)

	// log.Printf("[zest] workspace launch completed successfully for '%s'", wspName)
	return nil
}
