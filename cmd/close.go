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
	"os"
	"time"

	"github.com/AVAniketh0905/zest/internal/workspace"
	"github.com/spf13/cobra"
)

// TODO: closeCmd represents the close command
var closeCmd = &cobra.Command{
	Use:   "close [workspace-name]",
	Short: "Close an existing or active workspace",
	Long:  `Close a specific workspace by name, or use --all to close all workspaces.`,
	Example: `  zest close work
  zest close --all                
  zest close work --force`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cmd.ValidateArgs(args); err != nil {
			return err
		}
		return closeWorkspace(args[0])
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// closeCmd.PersistentFlags().String("foo", "", "A help for foo")

	closeCmd.Flags().BoolP("force", "f", false, "Force close the workspace even if it's active or has unsaved changes")
	closeCmd.Flags().Bool("all", false, "Close all currently open workspaces")
}

func closeWorkspace(wspName string) error {
	// log.Printf("[zest] closing sequence for workspace: %s", wspName)

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

	if wspCfg.Status == workspace.Inactive {
		// log.Printf("[zest] workspace '%s' is inactive", wspName)
		return workspace.ErrWorkspaceIsInactive
	}

	wspRt, err := workspace.NewWspRuntime(wspCfg.Name)
	if err != nil {
		// log.Printf("[zest] failed to initialize workspace runtime: %v", err)
		return err
	}
	// log.Printf("[zest] initialized workspace runtime")

	// TODO: kill all process in this workspace
	// BUG: not closing process
	for _, pid := range wspRt.PIDs {
		if err := killProcess(pid); err != nil {
			// log.Printf("[zest] failed to kill process, %d, %v", pid, err)
			return err
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

func killProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	return process.Kill()
}
