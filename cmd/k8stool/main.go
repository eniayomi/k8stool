package main

import (
	"log"
	"os"

	"github.com/eniayomi/k8stool/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}
} 