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
	"os"

	"github.com/AVAniketh0905/zest/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	VERSION = "0.1.0"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd *cobra.Command

// config file
var cfgFile string

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

}

// placeholder for all subcommands
func addCommands(rootCmd *cobra.Command, cfg *utils.ZestConfig) {
	rootCmd.AddCommand(NewInitCmd(cfg))
	rootCmd.AddCommand(NewStatusCmd(cfg))
	rootCmd.AddCommand(NewListCmd(cfg))
	rootCmd.AddCommand(NewLaunchCmd(cfg))
	rootCmd.AddCommand(NewCloseCmd(cfg))
}

func NewRootCmd(cfg *utils.ZestConfig) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "zest",
		Short: "Manage multiple workspaces from a unified CLI",
		Long: `Zest is a command-line tool to manage different workspaces with a simple
interface across platforms.

You can create and switch between isolated workspaces such as work, personal, or learning.
Each workspace can be initialized with custom templates for different use cases.`,
		Version: VERSION,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initConfig(cfg, cfgFile)
		},
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.zest/zest.yaml)")

	// custom path to zest directory
	rootCmd.PersistentFlags().StringVar(&cfg.ZestDir, "custom", "", "custom zest directory (default is $HOME/.zest)")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// Version
	versionTemplate := `{{printf "%s: %s - version %s\n" .Name .Short .Version}}`
	rootCmd.SetVersionTemplate(versionTemplate)

	// initialization
	addCommands(rootCmd, cfg)

	return rootCmd
}

func init() {
	cfg := &utils.ZestConfig{}
	RootCmd = NewRootCmd(cfg)
}

// initConfig reads in config file and ENV variables if set.
func initConfig(cfg *utils.ZestConfig, cfgFile string) {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		cfgPath := cfg.RootDir()

		// Zest configuration and state directory structure:
		//
		// $HOME/.zest/
		// ├── zest.yaml                        // Global configuration file (user editable)
		// ├── workspace/                       // Per-workspace configuration (user editable)
		// │   ├── [name of wsp].yaml           // Config for each workspace
		// ├── state/                           // Internal state files (NOT user editable)
		// │   ├── workspace.json               // Overall state of all workspaces
		// │   ├── other_future_cmds.json       // Additional future commands/state
		// │   └── workspace/                   // Per-workspace state files
		// │       ├── [name of wsp].json       // State for each workspace
		// Check for necessary directories.
		cobra.CheckErr(cfg.EnsureDirs())

		// config path at $HOME/.zest/zest.yaml
		viper.AddConfigPath(cfgPath)
		viper.SetConfigType("yaml")
		viper.SetConfigName("zest")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
