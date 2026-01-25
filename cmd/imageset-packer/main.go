package main

import (
	"os"

	"github.com/woozymasta/imageset-packer/internal/cli"
)

func main() {
	if err := cli.Run(os.Args[1:]); err != nil {
		// fmt.Fprintf(os.Stderr, "Error 1: %v\n", err)
		os.Exit(1)
	}
}
