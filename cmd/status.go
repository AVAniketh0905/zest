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
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/AVAniketh0905/zest/internal/utils"
	"github.com/AVAniketh0905/zest/internal/workspace"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

// statusCmd represents the status command
func NewStatusCmd(cfg *utils.ZestConfig) *cobra.Command {
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
 zest status --json
 zest status --since 1h
 zest status --watch`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.ValidateArgs(args); err != nil {
				return err
			}

			// log.Println("[zest] status: initializing workspace registry")
			wspReg, err := workspace.NewWspRegistry(cfg)
			if err != nil {
				return err
			}

			// filter args for proper workspaces
			workspaces, skipped, err := filterArgs(wspReg.GetNames(), args)
			if err != nil {
				return err
			}

			inactiveCh := make(chan *workspace.WspConfig, 3) // BUG: deadlock if not buffere
			activeCh := make(chan *workspace.WspRuntime, 3)

			wg := sync.WaitGroup{}
			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			defer func() {
				// log.Println("[zest] status: flushing tabwriter")
				tw.Flush()
			}()

			// Skipped Args
			if len(args) != 0 {
				fmt.Fprintln(tw, skipped)
			}

			// Writer goroutine
			wg.Add(1)
			go writerFunc(&wg, tw, inactiveCh, activeCh)

			// goroutine for populating active and inactive channels
			var g errgroup.Group

			// Launch one goroutine per workspace
			for _, wsp := range workspaces {
				g.Go(func() error {
					// log.Printf("[zest] status: checking workspace: %s", wsp)
					return getWspData(cfg, wspReg, wsp, inactiveCh, activeCh)
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

			return nil
		},
	}

	statusCmd.Flags().Bool("json", false, "Output in JSON format")
	statusCmd.Flags().BoolP("verbose", "v", false, "Show full runtime details")
	statusCmd.Flags().String("since", "", "Show only workspaces active since a given time")
	statusCmd.Flags().Bool("watch", false, "Watch live status with periodic updates")

	return statusCmd
}

func filterArgs(original, args []string) ([]string, string, error) {
	filter := []string{}

	origMap := make(map[string]bool)
	for _, item := range original {
		origMap[item] = true
	}

	skipped := "Skipped "
	for _, arg := range args {
		if _, ok := origMap[arg]; ok {
			filter = append(filter, arg)
		} else {
			skipped += arg + " "
		}
	}

	if skipped == "Skipped " {
		skipped += "None"
	}

	if len(filter) != 0 {
		return filter, skipped, nil
	}

	return original, "Skipped *", nil
}

func writerFunc(wg *sync.WaitGroup, tw *tabwriter.Writer, inactiveCh <-chan *workspace.WspConfig, activeCh <-chan *workspace.WspRuntime) {
	defer wg.Done()
	// log.Println("[zest] status: starting writer goroutine")

	// Inactive section
	fmt.Fprintln(tw, "INACTIVE")
	fmt.Fprintln(tw, "NAME\tSTATUS\tLAST_USED\tPATH")

	for inactive := range inactiveCh {
		// log.Printf("[zest] status: received inactive workspace: %s", inactive.Name)
		fmt.Fprintf(tw, "%s\tInactive\t%s\t%s\n",
			inactive.Name, inactive.LastUsed, inactive.Path)
	}

	// Active section
	fmt.Fprintln(tw, "ACTIVE")
	fmt.Fprintln(tw, "NAME\tSTATUS\tSTARTED_AT\tPIDS")

	for active := range activeCh {
		// log.Printf("[zest] status: received active workspace: %s", active.Name)
		pidStrSlice := []string{}
		for _, pid := range active.PIDs {
			pidStrSlice = append(pidStrSlice, fmt.Sprint(pid))
		}
		fmt.Fprintf(tw, "%s\tActive\t%s\t%s\n",
			active.Name, active.StartedAt, strings.Join(pidStrSlice, ","))
	}

	// log.Println("[zest] status: writer goroutine finished")
}

func getWspData(cfg *utils.ZestConfig, wspReg *workspace.WspRegistry, wsp string, inactiveCh chan<- *workspace.WspConfig, activeCh chan<- *workspace.WspRuntime) error {
	wspCfg, ok := wspReg.GetCfg(wsp)
	if !ok || cfg == nil {
		return fmt.Errorf("workspace %q not found or not initialized", wsp)
	}

	if wspCfg.Status == workspace.Inactive {
		// log.Printf("[zest] status: workspace %q is inactive", wsp)
		inactiveCh <- wspCfg
		return nil
	}

	// log.Printf("[zest] status: workspace %q is active, loading runtime", wsp)
	wspRt, err := workspace.NewWspRuntime(cfg, wsp)
	if err != nil {
		return err
	}
	if err := wspRt.Load(); err != nil {
		return fmt.Errorf("failed to load runtime for %q: %v", wsp, err)
	}

	// log.Printf("[zest] status: sending runtime for %q to channel", wsp)
	activeCh <- wspRt
	return nil
}
