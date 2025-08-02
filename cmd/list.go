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

// listCmd represents the list command
func NewListCmd(cfg *utils.ZestConfig) *cobra.Command {
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all available workspaces",
		Long: `List all workspaces that have been initialized using 'zest init'.

This command reads from the internal workspace registry and displays metadata
about each known workspace, such as its name, current status (open or closed),
last used timestamp, and configuration file path.`,
		Example: ` zest list
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

			wspData, err := getWorkspaces(cfg, strings.ToLower(filter), strings.ToLower(sortBy))
			if err != nil {
				return err
			}

			if isJson, err := cmd.Flags().GetBool("json"); err == nil && isJson {
				return renderWorkspacesJSON(cmd.OutOrStdout(), wspData)
			} else if err != nil {
				return err
			}

			return renderWorkspacesTable(cmd.OutOrStdout(), wspData)
		},
	}
	listCmd.Flags().Bool("json", false, "Output in JSON format")
	listCmd.Flags().String("filter", "all", "Filter by status: active, inactive, all")
	listCmd.Flags().String("sort", "name", "Sort by: name, last_used, status")

	return listCmd
}

type WorkspaceView struct {
	Name     string
	Status   workspace.Status
	LastUsed string
	Path     string
}

func sortWsp(s []WorkspaceView, sortBy string) ([]WorkspaceView, error) {
	switch sortBy {
	case "name":
		sort.Slice(s, func(i, j int) bool {
			return strings.ToLower(s[i].Name) < strings.ToLower(s[j].Name)
		})
	case "last_used":
		sort.Slice(s, func(i, j int) bool {
			return s[i].LastUsed < s[j].LastUsed // string sort (or parse time if needed)
		})
	case "status":
		sort.Slice(s, func(i, j int) bool {
			return strings.ToLower(string(s[i].Status)) < strings.ToLower(string(s[j].Status))
		})
	case "":
	default:
		return nil, fmt.Errorf("invalid sort key: %s (must be name, last_used, or status)", sortBy)
	}

	return s, nil
}

func getWorkspaces(cfg *utils.ZestConfig, filter string, sortBy string) ([]WorkspaceView, error) {
	wspReg, err := workspace.NewWspRegistry(cfg)
	if err != nil {
		return nil, err
	}

	var result []WorkspaceView
	for _, wsp := range wspReg.Workspaces {
		if filter != "all" && wsp.Status != workspace.Status(filter) {
			continue
		}

		_, trimPath, found := strings.Cut(wsp.Path, ".")
		if found {
			trimPath = filepath.Join("~", "."+trimPath)
		}

		result = append(result, WorkspaceView{
			Name:     wsp.Name,
			Status:   wsp.Status,
			LastUsed: wsp.LastUsed,
			Path:     trimPath,
		})
	}

	return sortWsp(result, sortBy)
}

func renderWorkspacesJSON(w io.Writer, workspaces []WorkspaceView) error {
	rawJson, err := json.MarshalIndent(workspaces, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal workspace data: %w", err)
	}

	_, err = w.Write(rawJson)
	return err
}

func renderWorkspacesTable(w io.Writer, workspaces []WorkspaceView) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSTATUS\tLAST_USED\tPATH")

	for _, wsp := range workspaces {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			wsp.Name,
			wsp.Status,
			wsp.LastUsed,
			wsp.Path,
		)
	}

	return tw.Flush()
}
