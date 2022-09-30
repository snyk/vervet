package main

import (
	"log"
	"os"

	"github.com/snyk/vervet/v5/internal/cmd"
)

func main() {
	err := cmd.Vervet.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
