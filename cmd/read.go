package cmd

import (
	"f5ltm/com"
	"f5ltm/pkg"
	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Read the excel configuration and load it to the f5 device",
	RunE: func(cmd *cobra.Command, args []string) error {
		return readRun()
	},
}

func readRun() error {
	client, _ := pkg.NewF5Clients()
	vs := pkg.NewVirtualServers()
	return vs.Read(client)
}

func init() {
	readCmd.PersistentFlags().StringVarP(&com.Host, "host", "a", "127.0.0.1", "The host ip address")
	readCmd.PersistentFlags().StringVarP(&com.Username, "username", "u", "admin", "The username of the host")
	readCmd.PersistentFlags().StringVarP(&com.Password, "password", "p", "admin", "Password for the given user")
	readCmd.PersistentFlags().StringVarP(&com.File, "file", "f", "/tmp/create.xlsx", "Specifies an alternative configuration file")
	readCmd.PersistentFlags().StringVarP(&com.Sheet, "sheet", "s", "Sheet1", "Specifies the table name of the workbook")

	readCmd.MarkFlagRequired(com.Host)
	readCmd.MarkFlagRequired(com.Username)
	readCmd.MarkFlagRequired(com.Password)
	readCmd.MarkFlagRequired(com.File)
	readCmd.MarkFlagRequired(com.Sheet)
	rootCmd.AddCommand(readCmd)
}
