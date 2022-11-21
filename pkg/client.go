package pkg

import (
	"fmt"
	"github.com/e-XpertSolutions/f5-rest-client/f5"
	"log"
)

func NewF5Client() (*f5.Client, error) {
	hosts := fmt.Sprintf("https://" + Host)
	client, err := f5.NewBasicClient(hosts, Username, Password)
	//clients, err := f5.NewBasicClient("https://192.168.10.84", "admin", "admin")
	client.DisableCertCheck()
	if err != nil {
		log.Fatalf("clients connect to f5 device failed: %s", err)
	}
	return client, nil
}
