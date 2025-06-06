package restore

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"yourproject/comfyui-compactor/oak"
)

// ExecuteRestore orchestrates the restoration of a ComfyUI installation from a snapshot.
func ExecuteRestore(snapshotFilePath string, targetDirectoryPath string) error {
	log.Printf("Starting restore process from snapshot: %s to target: %s", snapshotFilePath, targetDirectoryPath)

	// 1. Read Snapshot
	log.Println("Reading snapshot file...")
	snapshotData, err := os.ReadFile(snapshotFilePath)
	if err != nil {
		return fmt.Errorf("failed to read snapshot file %s: %w", snapshotFilePath, err)
	}

	var snapshot oak.Snapshot
	if err := json.Unmarshal(snapshotData, &snapshot); err != nil {
		return fmt.Errorf("failed to unmarshal snapshot JSON from %s: %w", snapshotFilePath, err)
	}
	log.Printf("Successfully read and parsed snapshot version: %s, Oak version: %s", snapshot.ProgramVersion, snapshot.OakVersion)

	// 2. Prepare Target Directory
	log.Printf("Preparing target directory: %s", targetDirectoryPath)
	fileInfo, err := os.Stat(targetDirectoryPath)
	if err == nil { // Path exists
		if !fileInfo.IsDir() {
			return fmt.Errorf("target path %s exists but is not a directory", targetDirectoryPath)
		}
		// Check if directory is empty
		dirEntries, err := os.ReadDir(targetDirectoryPath)
		if err != nil {
			return fmt.Errorf("failed to read target directory %s: %w", targetDirectoryPath, err)
		}
		if len(dirEntries) > 0 {
			return fmt.Errorf("target directory %s exists and is not empty. Please provide an empty or non-existent directory", targetDirectoryPath)
		}
		log.Printf("Target directory %s exists and is empty.", targetDirectoryPath)
	} else if os.IsNotExist(err) { // Path does not exist
		log.Printf("Target directory %s does not exist. Creating...", targetDirectoryPath)
		if err := os.MkdirAll(targetDirectoryPath, 0755); err != nil {
			return fmt.Errorf("failed to create target directory %s: %w", targetDirectoryPath, err)
		}
		log.Printf("Successfully created target directory %s", targetDirectoryPath)
	} else { // Other error (e.g., permission issues)
		return fmt.Errorf("failed to stat target directory %s: %w", targetDirectoryPath, err)
	}

	// 3. Instantiate Restorers
	comfyRestorer := ComfyUIRestorer{}
	nodeRestorer := NodeRestorer{}
	log.Println("ComfyUIRestorer and NodeRestorer instantiated.")

	// 4. Restore ComfyUI Core
	log.Println("Calling ComfyUI core restorer...")
	if err := comfyRestorer.Restore(targetDirectoryPath, snapshot.ComfyUI); err != nil {
		return fmt.Errorf("failed to restore ComfyUI core: %w", err)
	}
	log.Println("ComfyUI core restoration complete.")

	// 5. Restore Custom Nodes
	log.Println("Calling Custom Nodes restorer...")
	if err := nodeRestorer.Restore(targetDirectoryPath, snapshot.Nodes); err != nil {
	 // The NodeRestorer.Restore method already accumulates errors and returns a summary.
		return fmt.Errorf("failed during custom nodes restoration: %w", err)
	}
	log.Println("Custom nodes restoration process finished.")

	log.Println("Restore process fully completed.")
	return nil
}