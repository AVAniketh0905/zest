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
its startup plan. You can use --env to inject environment variables, --dry-run to preview,
or --detach to run in background.`,
		Example: `  zest launch work
  zest launch work --detach
  zest launch personal --dry-run
  zest launch work --env MODE=dev
  zest launch work --force
  zest launch personal --dry-run --env MODE=test`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			wspName := args[0]

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

			return launchWorkspace(cmd.OutOrStdout(), cfg, opts, wspName)
		},
	}

	launchCmd.Flags().Bool("dry-run", false, "Validate config and simulate launch without executing")
	launchCmd.Flags().BoolP("detach", "d", false, "Run workspace in background")
	launchCmd.Flags().StringToString("env", nil, "Set or override environment variables (e.g. --env KEY=VALUE)")
	launchCmd.Flags().BoolP("force", "f", false, "Force launch even if workspace is active")

	return launchCmd
}

func launchWorkspace(w io.Writer, cfg *utils.ZestConfig, opts LaunchOptions, wspName string) error {
	fmt.Fprintf(w, "Launching workspace '%s'...\n", wspName)

	wspReg, err := workspace.NewWspRegistry(cfg)
	if err != nil {
		return fmt.Errorf("unable to load workspace registry: %w", err)
	}

	wspCfg, ok := wspReg.GetCfg(wspName)
	if !ok {
		return fmt.Errorf("workspace '%s' does not exist", wspName)
	}

	// Block if already active and not forcing
	if wspCfg.Status == workspace.Active && !opts.Force {
		fmt.Fprintf(w, "Workspace '%s' is already active. Use --force to re-launch.\n", wspName)
		return workspace.ErrWorkspaceIsActive
	}

	// Build launch plan
	plan, err := launch.NewLaunchPlan(cfg, wspName)
	if err != nil {
		return fmt.Errorf("failed to create launch plan for '%s': %w", wspName, err)
	}

	if len(opts.Env) > 0 {
		fmt.Fprintf(w, "Applying environment variables: %v\n", opts.Env)
		plan.ApplyEnv(opts.Env)
	}

	wspRt, err := workspace.NewWspRuntime(cfg, wspCfg.Name)
	if err != nil {
		return fmt.Errorf("failed to initialize runtime for '%s': %w", wspName, err)
	}

	// Dry run mode
	if opts.DryRun {
		fmt.Fprintln(w, "[zest] Dry-run mode enabled. Launch plan preview:")
		fmt.Fprintln(w, plan.Summary())
		return nil
	}

	// Start execution of the plan
	fmt.Fprintln(w, "Starting launch...")
	if err := plan.Start(); err != nil {
		return fmt.Errorf("failed to launch workspace '%s': %w", wspName, err)
	}

	// Update and persist runtime state
	wspRt.Update(plan)
	if err := wspRt.Save(); err != nil {
		return fmt.Errorf("failed to save runtime state: %w", err)
	}

	// Mark workspace as active
	wspCfg.Status = workspace.Active
	wspReg.Update(wspCfg)
	if err := wspReg.Save(); err != nil {
		return fmt.Errorf("failed to update registry: %w", err)
	}

	fmt.Fprintf(w, "Workspace '%s' launched successfully.\n", wspName)
	return nil
}
