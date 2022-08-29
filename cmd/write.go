package cmd

import (
	"f5ltm/com"
	"f5ltm/pkg"
	"github.com/spf13/cobra"
)

var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "Read f5 device data and write to excel sheet",
	RunE: func(cmd *cobra.Command, args []string) error {
		return writeRun()
	},
}

func writeRun() error {
	client, _ := pkg.NewF5Clients()
	vs := pkg.NewVirtualServer()
	return vs.Write(client)
}

func init() {
	writeCmd.PersistentFlags().StringVarP(&com.Host, "host", "a", "127.0.0.1", "The host ip address")
	writeCmd.PersistentFlags().StringVarP(&com.Username, "username", "u", "admin", "The username of the host")
	writeCmd.PersistentFlags().StringVarP(&com.Password, "password", "p", "admin", "Password for the given user")
	writeCmd.PersistentFlags().StringVarP(&com.File, "file", "f", "/tmp/create.xlsx", "Specifies an alternative configuration file")
	writeCmd.PersistentFlags().StringVarP(&com.Sheet, "sheet", "s", "Sheet1", "Specifies the table name of the workbook")

	writeCmd.MarkFlagRequired(com.Host)
	writeCmd.MarkFlagRequired(com.Username)
	writeCmd.MarkFlagRequired(com.Password)
	writeCmd.MarkFlagRequired(com.File)
	writeCmd.MarkFlagRequired(com.Sheet)
	rootCmd.AddCommand(writeCmd)
}
