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
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/AVAniketh0905/zest/internal/utils"
	"github.com/AVAniketh0905/zest/internal/workspace"
	"github.com/spf13/cobra"
)

type WorkspaceView struct {
	Name     string
	Status   workspace.Status
	LastUsed string
	Path     string
}

// listCmd represents the list command
func NewListCmd(cfg *utils.ZestConfig) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all available workspaces",
		Long: `Displays all initialized workspaces, showing status, last-used time, and path.

Supports optional filtering (active/inactive), sorting (name, status, last_used),
and JSON output for automation or scripting.`,
		Example: `  zest list
  zest list --json
  zest list --filter active
  zest list --sort last_used`,
		RunE: func(cmd *cobra.Command, args []string) error {
			filter, err := cmd.Flags().GetString("filter")
			if err != nil {
				return err
			}
			sortBy, err := cmd.Flags().GetString("sort")
			if err != nil {
				return err
			}
			isJSON, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

			// Load and filter workspaces
			wspData, err := getWorkspaces(cfg, strings.ToLower(filter), strings.ToLower(sortBy))
			if err != nil {
				return err
			}

			if len(wspData) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No workspaces found.")
				return nil
			}

			// Render output
			if isJSON {
				return renderWorkspacesJSON(cmd.OutOrStdout(), wspData)
			}
			return renderWorkspacesTable(cmd.OutOrStdout(), wspData)
		},
	}

	// Flags
	listCmd.Flags().Bool("json", false, "Output in JSON format")
	listCmd.Flags().String("filter", "all", "Filter by status: active, inactive, all")
	listCmd.Flags().String("sort", "name", "Sort by: name, last_used, status")

	return listCmd
}

func sortWsp(s []WorkspaceView, sortBy string) ([]WorkspaceView, error) {
	switch sortBy {
	case "name":
		sort.Slice(s, func(i, j int) bool {
			return strings.ToLower(s[i].Name) < strings.ToLower(s[j].Name)
		})
	case "last_used":
		sort.Slice(s, func(i, j int) bool {
			return s[i].LastUsed < s[j].LastUsed // can improve with parsed times
		})
	case "status":
		sort.Slice(s, func(i, j int) bool {
			return strings.ToLower(string(s[i].Status)) < strings.ToLower(string(s[j].Status))
		})
	case "":
		return s, nil
	default:
		return nil, fmt.Errorf("invalid sort key: '%s' (expected: name, last_used, or status)", sortBy)
	}
	return s, nil
}

func getWorkspaces(cfg *utils.ZestConfig, filter string, sortBy string) ([]WorkspaceView, error) {
	wspReg, err := workspace.NewWspRegistry(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load workspace registry: %w", err)
	}

	var result []WorkspaceView
	for _, wsp := range wspReg.Workspaces {
		if filter != "all" && string(wsp.Status) != filter {
			continue
		}

		// Display-friendly path trimming
		_, trimmed, found := strings.Cut(wsp.Path, ".")
		if found {
			trimmed = filepath.Join("~", "."+trimmed)
		} else {
			trimmed = wsp.Path
		}

		result = append(result, WorkspaceView{
			Name:     wsp.Name,
			Status:   wsp.Status,
			LastUsed: wsp.LastUsed,
			Path:     trimmed,
		})
	}

	sorted, err := sortWsp(result, sortBy)
	if err != nil {
		return nil, err
	}

	return sorted, nil
}

func renderWorkspacesJSON(w io.Writer, workspaces []WorkspaceView) error {
	data, err := json.MarshalIndent(workspaces, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode workspace data to JSON: %w", err)
	}
	_, err = w.Write(data)
	return err
}

func renderWorkspacesTable(w io.Writer, workspaces []WorkspaceView) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSTATUS\tLAST USED\tPATH")

	for _, wsp := range workspaces {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			wsp.Name,
			strings.ToUpper(string(wsp.Status)),
			wsp.LastUsed,
			wsp.Path,
		)
	}

	return tw.Flush()
}
