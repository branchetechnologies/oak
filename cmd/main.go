package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
	restore "yourproject/comfyui-compactor/cmd/restore"
	snapshot "yourproject/comfyui-compactor/cmd/snapshot"
	// Note: The direct imports for oak, oak/restore, oak/snapshot are no longer needed here
	// as that logic is encapsulated within the respective cmd packages.
)

func main() {
	app := &cli.App{
		Name:  "oak",
		Usage: "A tool for ComfyUI snapshot and restore operations",
		Commands: []*cli.Command{
			snapshot.Command(),
			restore.Command(),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}