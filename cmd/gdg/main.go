package main

import (
	"log"
	"os"

	"github.com/esnet/gdg/cli"
)

func main() {
	err := cli.Execute(os.Args[1:])
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}
