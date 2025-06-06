package snapshotcmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/urfave/cli/v2" // Import urfave/cli
	"yourproject/comfyui-compactor/oak" // Import the local oak package for oak.Snapshot type
	"yourproject/comfyui-compactor/oak/snapshot" // Import the snapshot logic
)

// Command exports the snapshot command.
func Command() *cli.Command {
	return &cli.Command{
		Name:      "snapshot",
		Usage:     "Create a snapshot of a ComfyUI installation",
		ArgsUsage: "<path/to/comfyui>",
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return cli.Exit("Usage: oak snapshot <path/to/comfyui>", 1)
			}
			comfyUIPath := c.Args().First()

			// Ensure comfyUIPath is an absolute path, as sub-functions might rely on this.
			absComfyUIPath, err := filepath.Abs(comfyUIPath)
			if err != nil {
				return cli.Exit(fmt.Sprintf("Error converting path to absolute: %v", err), 1)
			}

			// Explicitly declare the type to ensure the oak package is used
			var snapData *oak.Snapshot
			snapData, err = snapshot.ExecuteSnapshot(absComfyUIPath)
			if err != nil {
				return cli.Exit(fmt.Sprintf("Error creating snapshot: %v", err), 1)
			}

			jsonData, err := json.MarshalIndent(snapData, "", "  ")
			if err != nil {
				return cli.Exit(fmt.Sprintf("Error marshalling snapshot to JSON: %v", err), 1)
			}

			fmt.Println(string(jsonData))
			return nil
		},
	}
}