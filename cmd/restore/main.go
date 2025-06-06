package restorecmd

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"yourproject/comfyui-compactor/oak/restore"
)

// Command exports the restore command.
func Command() *cli.Command {
	return &cli.Command{
		Name:      "restore",
		Usage:     "Restore a ComfyUI installation from a snapshot",
		ArgsUsage: "<path/to/snapshot.json> <path/to/target_comfyui_directory>",
		Action: func(c *cli.Context) error {
			if c.NArg() != 2 {
				return cli.Exit("Usage: oak restore <path/to/snapshot.json> <path/to/target_comfyui_directory>", 1)
			}
			snapshotFilePath := c.Args().Get(0)
			targetDirectoryPath := c.Args().Get(1)

			absSnapshotFilePath, err := filepath.Abs(snapshotFilePath)
			if err != nil {
				return cli.Exit(fmt.Sprintf("Error converting snapshot path to absolute: %v", err), 1)
			}

			absTargetDirectoryPath, err := filepath.Abs(targetDirectoryPath)
			if err != nil {
				return cli.Exit(fmt.Sprintf("Error converting target directory path to absolute: %v", err), 1)
			}

			log.Printf("Snapshot file: %s", absSnapshotFilePath)
			log.Printf("Target directory: %s", absTargetDirectoryPath)

			err = restore.ExecuteRestore(absSnapshotFilePath, absTargetDirectoryPath)
			if err != nil {
				return cli.Exit(fmt.Sprintf("Error during restore operation: %v", err), 1)
			}

			log.Println("Restore operation completed successfully.")
			fmt.Println("Restore operation completed successfully. See logs for details.")
			return nil
		},
	}
}