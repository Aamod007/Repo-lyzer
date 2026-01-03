package cmd

import "github.com/spf13/cobra"

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze a single repository",
	Run: func(cmd *cobra.Command, args []string) {
		// Placeholder
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}
