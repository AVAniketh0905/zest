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
	"github.com/AVAniketh0905/zest/internal/launch"
	"github.com/AVAniketh0905/zest/internal/utils"
	"github.com/AVAniketh0905/zest/internal/workspace"
	"github.com/spf13/cobra"
)

// launchCmd represents the launch command
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
  zest launch personal --dry-run --env MODE=test --cmd "./custom-start.sh"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cmd.ValidateArgs(args); err != nil {
			return err
		}
		return launchWorkspace(args[0])
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// launchCmd.PersistentFlags().String("foo", "", "A help for foo")

	launchCmd.Flags().Bool("dry-run", false, "Validate config and simulate launch without executing")
	launchCmd.Flags().BoolP("detach", "d", false, "Run workspace in background")
	launchCmd.Flags().StringToString("env", nil, "Set or override environment variables (e.g. --env KEY=VALUE)")
	launchCmd.Flags().BoolP("force", "f", false, "Force launch even if workspace is active")
}

func launchWorkspace(wspName string) error {
	// log.Printf("[zest] starting launch sequence for workspace: %s", wspName)

	wspReg, err := workspace.NewWspRegistry()
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

	if wspCfg.Status == workspace.Active {
		// log.Printf("[zest] workspace '%s' is already active", wspName)
		return workspace.ErrWorkspaceIsActive
	}
	// log.Printf("[zest] workspace '%s' is valid and inactive", wspName)

	plan, err := launch.NewLaunchPlan(wspName)
	if err != nil {
		// log.Printf("[zest] failed to parse launch plan: %v", err)
		return err
	}
	// log.Printf("[zest] parsed launch plan for workspace '%s'", wspName)

	wspRt, err := workspace.NewWspRuntime(wspCfg.Name)
	if err != nil {
		// log.Printf("[zest] failed to initialize workspace runtime: %v", err)
		return err
	}
	// log.Printf("[zest] initialized workspace runtime")

	if err := plan.Start(); err != nil {
		// log.Printf("[zest] failed to launch apps: %v", err)
		return err
	}
	// log.Printf("[zest] launched apps successfully")

	wspRt.Update(plan)
	// log.Printf("[zest] updated workspace runtime state")

	// update pids with actual process pids
	for i, pname := range wspRt.Processes {
		after, err := utils.ListPIDs(pname)
		if err != nil {
			return err
		}
		wspRt.PIDs[i] = utils.Diff(after, wspRt.PIDs[i])
	}

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
