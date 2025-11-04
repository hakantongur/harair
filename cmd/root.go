package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	cfgPath   string
	rulesPath string
	verbose   bool
)

var rootCmd = &cobra.Command{
	Use:   "harair",
	Short: "Harbor Air-Gap CLI",
	Long:  `harair: Mirror, export, and import Harbor images & Helm charts across air-gapped networks.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			color.New(color.FgCyan).Println("[harair] verbose logging enabled")
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "config.go", "Path to config.go")
	rootCmd.PersistentFlags().StringVar(&rulesPath, "rules", "rules.yaml", "Path to rules.yaml")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
}
