package main

import (
	"log"
	"os"

	"github.com/snyk/vervet/cmd"
)

func main() {
	err := cmd.Vervet.App.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
