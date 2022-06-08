package main

import (
	"f5ltm/ltm"
	"log"
)

func main() {
	client, _ := ltm.NewF5Client()
	vs := ltm.NewVirtualServers()
	if err := vs.Exec(client); err != nil {
		log.Fatal(err)
	}
}
