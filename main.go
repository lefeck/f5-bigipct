package main

import (
	"f5ltm/cmd"
	"log"
	"os"
)

func main() {
	app := cmd.Init()
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
