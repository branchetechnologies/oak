package restore

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"yourproject/comfyui-compactor/oak/functions"
)

// ComfyUIRestorer handles the restoration of the ComfyUI core.
type ComfyUIRestorer struct{}

// Restore clones the ComfyUI repository and checks out the specified commit.
// The targetDirectoryPath is the path where ComfyUI core should be cloned into.
func (cr *ComfyUIRestorer) Restore(targetDirectoryPath string, comfyUIGitInfo functions.GitInfo) error {
	log.Printf("Restoring ComfyUI core to %s", targetDirectoryPath)

	if comfyUIGitInfo.GitURL == "" || comfyUIGitInfo.GitURL == "unknown" {
		return fmt.Errorf("ComfyUI GitURL is missing or unknown in snapshot")
	}
	if comfyUIGitInfo.CommitSHA == "" || comfyUIGitInfo.CommitSHA == "unknown" {
		return fmt.Errorf("ComfyUI CommitSHA is missing or unknown in snapshot")
	}

	// Parent directory of the target. Git clone will create the last part of targetDirectoryPath.
	parentDir := filepath.Dir(targetDirectoryPath)
	cloneDirName := filepath.Base(targetDirectoryPath)

	// Ensure parent directory exists (it should, as targetDirectoryPath was either created or verified empty)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		// This case should ideally not be hit if ExecuteRestore prepared targetDirectoryPath correctly.
		// However, creating it defensively.
		if errMk := os.MkdirAll(parentDir, 0755); errMk != nil {
			return fmt.Errorf("failed to create parent directory %s for clone: %w", parentDir, errMk)
		}
	}


	log.Printf("Cloning ComfyUI from %s into %s/%s...", comfyUIGitInfo.GitURL, parentDir, cloneDirName)
	_, err := functions.RunCommand(parentDir, "git", "clone", comfyUIGitInfo.GitURL, cloneDirName)
	if err != nil {
		return fmt.Errorf("failed to clone ComfyUI repository from %s: %w", comfyUIGitInfo.GitURL, err)
	}
	log.Printf("Successfully cloned ComfyUI to %s", targetDirectoryPath)

	log.Printf("Checking out ComfyUI commit %s in %s...", comfyUIGitInfo.CommitSHA, targetDirectoryPath)
	_, err = functions.RunCommand(targetDirectoryPath, "git", "checkout", comfyUIGitInfo.CommitSHA)
	if err != nil {
		// Attempt to fetch before checkout if checkout fails, as the commit might not be locally available
		// especially if the default branch was shallow cloned.
		log.Printf("Checkout failed, attempting to fetch all and retry checkout for commit %s...", comfyUIGitInfo.CommitSHA)
		_, fetchErr := functions.RunCommand(targetDirectoryPath, "git", "fetch", "--all", "--tags")
		if fetchErr != nil {
			log.Printf("Warning: 'git fetch --all --tags' also failed: %v", fetchErr)
			// Return original checkout error
			return fmt.Errorf("failed to checkout ComfyUI commit %s after clone (and fetch attempt failed): %w", comfyUIGitInfo.CommitSHA, err)
		}
		_, err = functions.RunCommand(targetDirectoryPath, "git", "checkout", comfyUIGitInfo.CommitSHA)
		if err != nil {
			return fmt.Errorf("failed to checkout ComfyUI commit %s even after fetch: %w", comfyUIGitInfo.CommitSHA, err)
		}
	}
	log.Printf("Successfully checked out ComfyUI commit %s", comfyUIGitInfo.CommitSHA)

	log.Println("ComfyUI core restoration complete.")
	return nil
}