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
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/AVAniketh0905/zest/internal/utils"
	"github.com/AVAniketh0905/zest/internal/workspace"
	"github.com/spf13/cobra"
)

type StatusReport struct {
	Skipped   []string                `json:"skipped,omitempty"`
	Inactive  []*workspace.WspConfig  `json:"inactive"`
	Active    []*workspace.WspRuntime `json:"active"`
	Timestamp string                  `json:"generated_at"`
}

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
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

			verbose, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			sinceStr, err := cmd.Flags().GetString("since")
			if err != nil {
				return err
			}

			watch, err := cmd.Flags().GetBool("watch")
			if err != nil {
				return err
			}

			since, err := parseSinceFlag(sinceStr)
			if err != nil {
				return err
			}

			runOnce := func() error {
				return runStatusOnce(cmd.OutOrStdout(), cfg, args, jsonOut, verbose, since)
			}

			if watch {
				return watchStatus(cmd, runOnce)
			}

			return runOnce()
		},
	}

	statusCmd.Flags().Bool("json", false, "Output in JSON format")
	statusCmd.Flags().BoolP("verbose", "v", false, "Show full runtime details")
	statusCmd.Flags().String("since", "", "Show only workspaces active since a given time")
	statusCmd.Flags().Bool("watch", false, "Watch live status with periodic updates")

	return statusCmd
}

func flatten[T comparable](s [][]T) []T {
	var flat []T
	for _, group := range s {
		flat = append(flat, group...)
	}
	return flat
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func joinInts(ints []int) string {
	if len(ints) == 0 {
		return "-"
	}
	var out []string
	for _, v := range ints {
		out = append(out, strconv.Itoa(v))
	}
	return strings.Join(out, ",")
}

func wrapEmptyOutput(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

func parseSinceFlag(sinceStr string) (time.Time, error) {
	var since time.Time
	if sinceStr != "" {
		dur, err := time.ParseDuration(sinceStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid duration for --since: %v", err)
		}
		since = time.Now().Add(-dur)
	}

	return since, nil
}

func filterArgs(available, args []string) ([]string, []string) {
	if len(args) == 0 {
		return available, nil
	}

	availableMap := make(map[string]bool)
	for _, name := range available {
		availableMap[name] = true
	}

	var valid []string
	var skipped []string
	for _, arg := range args {
		if availableMap[arg] {
			valid = append(valid, arg)
		} else {
			skipped = append(skipped, arg)
		}
	}

	return valid, skipped
}

func renderJSON(w io.Writer, actives []*workspace.WspRuntime, inactives []*workspace.WspConfig, skipped []string) error {
	report := StatusReport{
		Skipped:   skipped,
		Inactive:  inactives,
		Active:    actives,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(w, string(data))
	return nil
}

func renderStatusTable(w io.Writer, actives []*workspace.WspRuntime, inactives []*workspace.WspConfig, skipped []string, verbose bool) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	if len(skipped) > 0 {
		fmt.Fprintf(tw, "Skipped: %s\n", strings.Join(skipped, ", "))
	}

	if len(inactives) > 0 {
		fmt.Fprintln(tw, "\nINACTIVE WORKSPACES")
		fmt.Fprintln(tw, "NAME\tSTATUS\tLAST_USED\tPATH")
		for _, wsp := range inactives {
			fmt.Fprintf(tw, "%s\tInactive\t%s\t%s\n", wsp.Name, wsp.LastUsed, wsp.Path)
		}
	}

	if len(actives) > 0 {
		fmt.Fprintln(tw, "\nACTIVE WORKSPACES")
		fmt.Fprintln(tw, "NAME\tSTATUS\tSTARTED_AT\tPIDS\tPROCESSES")

		for _, wsp := range actives {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
				wsp.Name,
				"Active",
				wsp.StartedAt,
				truncate(joinInts(flatten(wsp.PIDs)), 30),
				wrapEmptyOutput(strings.Join(wsp.Processes, ",")),
			)
		}
		tw.Flush()

		// Add space before verbose details
		if verbose {
			fmt.Fprintln(w, "\nExtra Details:")
			for _, wsp := range actives {
				fmt.Fprintf(w, "%s:\n", wsp.Name)
				flatPids := flatten(wsp.PIDs)
				if len(flatPids) > 6 {
					fmt.Fprintf(w, "  PIDs: %s\n", joinInts(flatPids))
				}
				if len(wsp.Ports) > 0 {
					fmt.Fprintf(w, "  Ports: %s\n", joinInts(wsp.Ports))
				}
				if len(wsp.BrowserURLs) > 0 {
					fmt.Fprintf(w, "  URLs:      %s\n", strings.Join(wsp.BrowserURLs, ", "))
				}
				fmt.Fprintf(w, "  Detached:  %v\n\n", wsp.IsDetached)
			}
		}
	} else {
		fmt.Fprintln(w, "No active workspaces.")
	}

	return tw.Flush()
}

func runStatusOnce(w io.Writer, cfg *utils.ZestConfig, args []string, jsonOut, verbose bool, since time.Time) error {
	wspReg, err := workspace.NewWspRegistry(cfg)
	if err != nil {
		return err
	}

	allWspNames := wspReg.GetNames()
	selected, skipped := filterArgs(allWspNames, args)

	var actives []*workspace.WspRuntime
	var inactives []*workspace.WspConfig

	for _, wsp := range selected {
		wspCfg, ok := wspReg.GetCfg(wsp)
		if !ok {
			continue
		}

		if wspCfg.Status == workspace.Inactive {
			inactives = append(inactives, wspCfg)
			continue
		}

		rt, err := workspace.NewWspRuntime(cfg, wsp)
		if err != nil {
			continue
		}
		if err := rt.Load(); err != nil {
			continue
		}

		if !since.IsZero() {
			started, err := time.Parse(time.RFC3339, rt.StartedAt)
			if err != nil || started.Before(since) {
				continue
			}
		}

		actives = append(actives, rt)
	}

	if jsonOut {
		return renderJSON(w, actives, inactives, skipped)
	}

	return renderStatusTable(w, actives, inactives, skipped, verbose)
}

func watchStatus(cmd *cobra.Command, runOnce func() error) error {
	for {
		fmt.Fprintf(cmd.OutOrStdout(), "Updated @ %s\n", time.Now().Format(time.Kitchen))
		if err := runOnce(); err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), "Error:", err)
		}
		time.Sleep(5 * time.Second)
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "\n---")
	}
}
