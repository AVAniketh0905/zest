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
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/AVAniketh0905/zest/internal/workspace"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status [workspace-name]",
	Short: "Show the live status of one or more workspaces",
	Long: `Displays runtime information about a specific workspace or all currently active workspaces.

This includes details such as which applications are running, associated process IDs,
container statuses, working directory, and other live session data tracked by Zest.

If no workspace name is provided, the status of all open workspaces will be shown.`,
	Example: ` zest status
 zest status personal
 zest status personal --verbose
 zest status --json`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cmd.ValidateArgs(args); err != nil {
			return err
		}

		if err := workspaceStatuses(cmd.OutOrStdout(), args); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func workspaceStatuses(w io.Writer, wsps []string) error {
	// log.Println("[zest] status: initializing workspace registry")

	wspReg, err := workspace.NewWspRegistry()
	if err != nil {
		return err
	}

	activeCh := make(chan *workspace.WspRuntime)
	inactiveCh := make(chan *workspace.WspConfig)

	wg := sync.WaitGroup{}

	// Writer goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		// log.Println("[zest] status: starting writer goroutine")

		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		defer func() {
			// log.Println("[zest] status: flushing tabwriter")
			tw.Flush()
		}()

		// Inactive section
		fmt.Fprintln(tw, workspace.Inactive)
		fmt.Fprintln(tw, "NAME\tSTATUS\tLAST_USED\tPATH")

		for inactive := range inactiveCh {
			// log.Printf("[zest] status: received inactive workspace: %s", inactive.Name)
			fmt.Fprintf(tw, "%s\tInactive\t%s\t%s\n",
				inactive.Name, inactive.LastUsed, inactive.Path)
		}

		// Active section
		fmt.Fprintln(tw, workspace.Active)
		fmt.Fprintln(tw, "NAME\tSTATUS\tSTARTED_AT\tPIDS")

		for active := range activeCh {
			// log.Printf("[zest] status: received active workspace: %s", active.Name)
			fmt.Fprintf(tw, "%s\tActive\t%s\t%s\n",
				active.Name, active.StartedAt, strings.Join(active.Processes, ","))
		}

		// log.Println("[zest] status: writer goroutine finished")
	}()

	// errgroup for writing to active and inactive channels
	var g errgroup.Group

	// Launch one goroutine per workspace
	for _, wsp := range wsps {
		g.Go(func() error {
			// log.Printf("[zest] status: checking workspace: %s", wsp)
			return wspStatus(wspReg, wsp, inactiveCh, activeCh)
		})
	}

	// Wait for all workers and the writer
	if err := g.Wait(); err != nil {
		// log.Printf("[zest] status: error encountered: %v", err)
		close(activeCh)
		close(inactiveCh)
		return err
	}

	// Close channels
	// log.Println("[zest] status: all workspace goroutines finished, closing channels")
	close(activeCh)
	close(inactiveCh)

	wg.Wait()

	// log.Println("[zest] status: completed successfully")
	return nil
}

func wspStatus(wspReg *workspace.WspRegistry, wsp string, inactiveCh chan<- *workspace.WspConfig, activeCh chan<- *workspace.WspRuntime) error {
	cfg, ok := wspReg.GetCfg(wsp)
	if !ok || cfg == nil {
		return fmt.Errorf("workspace %q not found or not initialized", wsp)
	}

	if cfg.Status == workspace.Inactive {
		// log.Printf("[zest] status: workspace %q is inactive", wsp)
		inactiveCh <- cfg
		return nil
	}

	// log.Printf("[zest] status: workspace %q is active, loading runtime", wsp)
	wspRt, err := workspace.NewWspRuntime()
	if err != nil {
		return fmt.Errorf("failed to load runtime for %q: %v", wsp, err)
	}

	// log.Printf("[zest] status: sending runtime for %q to channel", wsp)
	activeCh <- wspRt
	return nil
}
