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
	statusCmd := &cobra.Command{
		Use:   "status [workspace-name]",
		Short: "Show the live status of one or more workspaces",
		Long: `Displays runtime information about one or all currently active workspaces.

This includes application runtime status, process IDs, open ports, and other session details.

If no workspace is specified, status is shown for all.`,
		Example: `  zest status
  zest status personal
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
	statusCmd.Flags().BoolP("verbose", "v", false, "Show detailed process/runtime info")
	statusCmd.Flags().String("since", "", "Only include workspaces active since given duration (e.g. 1h, 30m)")
	statusCmd.Flags().Bool("watch", false, "Continuously watch and refresh status output")

	return statusCmd
}

func parseSinceFlag(val string) (time.Time, error) {
	if val == "" {
		return time.Time{}, nil
	}
	dur, err := time.ParseDuration(val)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid --since duration: %v", err)
	}
	return time.Now().Add(-dur), nil
}

func filterArgs(available, input []string) ([]string, []string) {
	if len(input) == 0 {
		return available, nil
	}
	availableMap := make(map[string]struct{})
	for _, a := range available {
		availableMap[a] = struct{}{}
	}
	var matched, skipped []string
	for _, arg := range input {
		if _, ok := availableMap[arg]; ok {
			matched = append(matched, arg)
		} else {
			skipped = append(skipped, arg)
		}
	}
	return matched, skipped
}

func joinInts(values []int) string {
	if len(values) == 0 {
		return "-"
	}
	var parts []string
	for _, v := range values {
		parts = append(parts, strconv.Itoa(v))
	}
	return strings.Join(parts, ",")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func flatten[T any](slices [][]T) []T {
	var flat []T
	for _, s := range slices {
		flat = append(flat, s...)
	}
	return flat
}

func wrapEmptyOutput(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

func watchStatus(cmd *cobra.Command, runOnce func() error) error {
	for {
		fmt.Fprintf(cmd.OutOrStdout(), "\nUpdated @ %s\n", time.Now().Format(time.Kitchen))
		if err := runOnce(); err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), "Error:", err)
		}
		time.Sleep(5 * time.Second)
		fmt.Fprintln(cmd.OutOrStdout(), "\n---")
	}
}

func runStatusOnce(w io.Writer, cfg *utils.ZestConfig, args []string, jsonOut, verbose bool, since time.Time) error {
	registry, err := workspace.NewWspRegistry(cfg)
	if err != nil {
		return fmt.Errorf("failed to load workspace registry: %w", err)
	}

	allNames := registry.GetNames()
	selected, skipped := filterArgs(allNames, args)

	var actives []*workspace.WspRuntime
	var inactives []*workspace.WspConfig

	for _, name := range selected {
		wspCfg, ok := registry.GetCfg(name)
		if !ok {
			continue
		}

		if wspCfg.Status == workspace.Inactive {
			inactives = append(inactives, wspCfg)
			continue
		}

		rt, err := workspace.NewWspRuntime(cfg, name)
		if err != nil || rt.Load() != nil {
			continue
		}

		// Filter by time, if applicable
		if !since.IsZero() {
			startedAt, err := time.Parse(time.RFC3339, rt.StartedAt)
			if err != nil || startedAt.Before(since) {
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

func renderJSON(w io.Writer, actives []*workspace.WspRuntime, inactives []*workspace.WspConfig, skipped []string) error {
	report := StatusReport{
		Skipped:   skipped,
		Inactive:  inactives,
		Active:    actives,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal status report: %w", err)
	}
	fmt.Fprintln(w, string(b))
	return nil
}

func renderStatusTable(w io.Writer, actives []*workspace.WspRuntime, inactives []*workspace.WspConfig, skipped []string, verbose bool) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	// Skipped
	if len(skipped) > 0 {
		fmt.Fprintf(tw, "Skipped: %s\n", strings.Join(skipped, ", "))
	}

	// Inactive
	if len(inactives) > 0 {
		fmt.Fprintln(tw, "\nINACTIVE WORKSPACES")
		fmt.Fprintln(tw, "NAME\tSTATUS\tLAST_USED\tPATH")
		for _, wsp := range inactives {
			fmt.Fprintf(tw, "%s\tInactive\t%s\t%s\n", wsp.Name, wsp.LastUsed, wsp.Path)
		}
	}

	// Active
	if len(actives) > 0 {
		fmt.Fprintln(tw, "\nACTIVE WORKSPACES")
		fmt.Fprintln(tw, "NAME\tSTATUS\tSTARTED_AT\tPIDS\tPROCESSES")
		for _, wsp := range actives {
			fmt.Fprintf(tw, "%s\tActive\t%s\t%s\t%s\n",
				wsp.Name,
				wsp.StartedAt,
				truncate(joinInts(flatten(wsp.PIDs)), 30),
				wrapEmptyOutput(strings.Join(wsp.Processes, ",")),
			)
		}
		tw.Flush()

		if verbose {
			fmt.Fprintln(w, "\nExtra Details:")
			for _, wsp := range actives {
				fmt.Fprintf(w, "\n%s:\n", wsp.Name)
				if len(flatten(wsp.PIDs)) > 6 {
					fmt.Fprintf(w, "  PIDs: %s\n", joinInts(flatten(wsp.PIDs)))
				}
				if len(wsp.Ports) > 0 {
					fmt.Fprintf(w, "  Ports: %s\n", joinInts(wsp.Ports))
				}
				if len(wsp.BrowserURLs) > 0 {
					fmt.Fprintf(w, "  URLs: %s\n", strings.Join(wsp.BrowserURLs, ", "))
				}
				fmt.Fprintf(w, "  Detached: %v\n", wsp.IsDetached)
			}
		}
	} else {
		fmt.Fprintln(w, "No active workspaces.")
	}

	return tw.Flush()
}
