package main

import (
	"log"
	"os"

	"github.com/snyk/vervet/v7/internal/cmd"
)

func main() {
	if err := cmd.Vervet.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
