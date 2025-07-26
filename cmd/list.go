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
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/AVAniketh0905/zest/internal/workspace"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available workspaces",
	Long: `List all workspaces that have been initialized using 'zest init'.

This command reads from the internal workspace registry and displays metadata
about each known workspace, such as its name, current status (open or closed),
last used timestamp, and configuration file path.`,
	Example: ` zest list
 zest list --json
 zest list --filter open
 zest list --sort last_used
 zest list --quiet`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listWorkspaces(cmd.OutOrStdout())
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	listCmd.Flags().Bool("json", false, "Output in JSON format")
	listCmd.Flags().String("filter", "all", "Filter by status: open, closed, all")
	listCmd.Flags().String("sort", "name", "Sort by: name, last_used, status")
	listCmd.Flags().BoolP("quiet", "q", false, "Only print workspace names")
}

func listWorkspaces(w io.Writer) error {
	wspReg, err := workspace.NewWspRegistry()
	if err != nil {
		return err
	}

	// Initialize a tab writer
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	// Print header
	fmt.Fprintln(tw, "NAME\tSTATUS\tLAST_USED\tPATH")

	// Loop and print each workspace
	for _, wsp := range wspReg.Workspaces {
		_, trimPath, found := strings.Cut(wsp.Path, ".")
		if found {
			trimPath = filepath.Join("~", "."+trimPath)
		}

		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			wsp.Name,
			wsp.Status,
			wsp.LastUsed,
			trimPath,
		)
	}

	// Flush to ensure output is written
	return tw.Flush()
}
