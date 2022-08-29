package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "ltm",
	Short: "ltm is a tool for reading and writing f5 bigip devices",
	Long: `ltm controls the f5 bigip devices.
  _____  ____   _    _____  __  __ 
 |  ___|| ___| | |  |_   _||  \/  |
 | |_   |___ \ | |    | |  | |\/| |
 |  _|   ___) || |___ | |  | |  | |
 |_|    |____/ |_____||_|  |_|  |_|
                                   
`,
	Run: runHelp,
}

func Executes() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runHelp(cmd *cobra.Command, args []string) {
	cmd.Help()
}
