package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var DemoCommand = &cobra.Command{
	Use:   "help",
	Short: "demo for framework",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("app base folder:")
	},
}
