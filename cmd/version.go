package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var version = "1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show ltm version",
	Run: func(c *cobra.Command, args []string) {
		fmt.Printf("ltm version %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
