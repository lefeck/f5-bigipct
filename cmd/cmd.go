package cmd

import (
	"f5ltm/pkg"
	"github.com/urfave/cli/v2"
	"log"
)

func NewApp() *cli.App {
	app := &cli.App{
		Name:                   "f5-bigipct",
		UseShortOptionHandling: true,
		Usage:                  "f5-bigipct controls the f5 bigip devices.",
		Version:                "2.0",
		Authors: []*cli.Author{
			{
				Name:  "Johnny Wilson",
				Email: "jw6759792@gmail.com",
			},
		},
	}

	app.Commands = []*cli.Command{
		{
			Name:        "import",
			Description: "Read the excel configuration and load it to the f5 device",
			Flags:       getFlag(),
			Action:      action,
		},
		{
			Name:        "export",
			Description: "Read f5 device data and write to excel sheet",
			Flags:       getFlag(),
			Action:      action,
		},
	}
	return app
}

func action(c *cli.Context) error {
	pkg.Host = c.String("host")
	pkg.Username = c.String("username")
	pkg.Password = c.String("password")
	pkg.File = c.String("file")
	pkg.Sheet = c.String("sheet")
	switch c.Command.Name {
	case "import":
		if err := imports(); err != nil {
			log.Fatalf("connect to failed: %s", err)
		}
	case "export":
		if err := exports(); err != nil {
			log.Fatalf("connect to failed: %s", err)
		}
	}
	return nil
}

func exports() error {
	client, _ := pkg.NewF5Client()
	vs := pkg.NewVirtualServer()
	return vs.Export(client)
}

func imports() error {
	client, _ := pkg.NewF5Client()
	vs := pkg.NewVirtualServers()
	return vs.Import(client)
}

func getFlag() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "host",
			Value:   "127.0.0.1",
			Usage:   "Enter the address to connect to the remote host.",
			Aliases: []string{"a"},
		},
		&cli.StringFlag{
			Name:    "username",
			Value:   "admin",
			Usage:   "Username to connect to the remote host.",
			Aliases: []string{"u"},
		},
		&cli.StringFlag{
			Name:    "password",
			Value:   "admin",
			Usage:   "Password to connect to the remote host.",
			Aliases: []string{"p"},
		},
		&cli.StringFlag{
			Name:    "file",
			Value:   "./ltm.xlsx",
			Usage:   "This file is used for read or write operations.",
			Aliases: []string{"f"},
		},
		&cli.StringFlag{
			Name:    "sheet",
			Value:   "Sheet1",
			Usage:   "The table name of the workbook.",
			Aliases: []string{"s"},
		},
	}
}
