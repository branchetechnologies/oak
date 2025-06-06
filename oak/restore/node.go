package restore

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"yourproject/comfyui-compactor/oak/functions"
)

// NodeRestorer handles the restoration of custom nodes.
type NodeRestorer struct{}

// Restore clones custom node repositories and checks out specified commits.
// targetDirectoryPath is the root of the ComfyUI installation.
func (nr *NodeRestorer) Restore(targetDirectoryPath string, nodes map[string]functions.NodeInfo) error {
	log.Println("Restoring custom nodes...")
	if len(nodes) == 0 {
		log.Println("No custom nodes found in snapshot to restore.")
		return nil
	}

	customNodesBasePath := filepath.Join(targetDirectoryPath, "custom_nodes")
	log.Printf("Ensuring custom_nodes directory exists at: %s", customNodesBasePath)
	if err := os.MkdirAll(customNodesBasePath, 0755); err != nil {
		return fmt.Errorf("failed to create custom_nodes directory %s: %w", customNodesBasePath, err)
	}

	var accumulatedErrors []string

	for nodeKey, nodeInfo := range nodes {
		log.Printf("Processing custom node: %s (Path: %s)", nodeKey, nodeInfo.Path)

		if nodeInfo.GitURL == "" || nodeInfo.GitURL == "unknown" {
			errMsg := fmt.Sprintf("Skipping node %s: GitURL is missing or unknown in snapshot.", nodeKey)
			log.Println(errMsg)
			accumulatedErrors = append(accumulatedErrors, errMsg)
			continue
		}
		// CommitSHA can be "unknown" if it's not a git repo, in which case we just clone.
		// If it's a git repo and commit is unknown, that's an issue, but GetGitInfo handles this.
		// For restoration, if CommitSHA is "unknown", we can only clone the default branch.

		// nodeInfo.Path is like "custom_nodes/NodeName". We need "NodeName".
		nodeDirName := filepath.Base(nodeInfo.Path)
		if nodeDirName == "." || nodeDirName == "custom_nodes" {
			errMsg := fmt.Sprintf("Skipping node %s: Invalid path in snapshot: %s", nodeKey, nodeInfo.Path)
			log.Println(errMsg)
			accumulatedErrors = append(accumulatedErrors, errMsg)
			continue
		}
		nodeTargetPath := filepath.Join(customNodesBasePath, nodeDirName)

		log.Printf("Cloning node %s from %s into %s...", nodeKey, nodeInfo.GitURL, nodeTargetPath)
		_, err := functions.RunCommand(customNodesBasePath, "git", "clone", nodeInfo.GitURL, nodeDirName)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to clone node %s from %s: %v", nodeKey, nodeInfo.GitURL, err)
			log.Println(errMsg)
			accumulatedErrors = append(accumulatedErrors, errMsg)
			continue // Move to the next node
		}
		log.Printf("Successfully cloned node %s to %s", nodeKey, nodeTargetPath)

		if nodeInfo.CommitSHA != "" && nodeInfo.CommitSHA != "unknown" {
			log.Printf("Checking out commit %s for node %s in %s...", nodeInfo.CommitSHA, nodeKey, nodeTargetPath)
			_, err = functions.RunCommand(nodeTargetPath, "git", "checkout", nodeInfo.CommitSHA)
			if err != nil {
				log.Printf("Checkout for node %s commit %s failed, attempting to fetch all and retry...", nodeKey, nodeInfo.CommitSHA)
				_, fetchErr := functions.RunCommand(nodeTargetPath, "git", "fetch", "--all", "--tags")
				if fetchErr != nil {
					log.Printf("Warning: 'git fetch --all --tags' for node %s also failed: %v", nodeKey, fetchErr)
				}
				_, err = functions.RunCommand(nodeTargetPath, "git", "checkout", nodeInfo.CommitSHA)
				if err != nil {
					errMsg := fmt.Sprintf("Failed to checkout commit %s for node %s (even after fetch): %v", nodeInfo.CommitSHA, nodeKey, err)
					log.Println(errMsg)
					accumulatedErrors = append(accumulatedErrors, errMsg)
					continue
				}
			}
			log.Printf("Successfully checked out commit %s for node %s", nodeInfo.CommitSHA, nodeKey)
		} else {
			log.Printf("Node %s: CommitSHA is '%s', cloned default branch.", nodeKey, nodeInfo.CommitSHA)
		}
	}

	if len(accumulatedErrors) > 0 {
		return fmt.Errorf("encountered errors during custom node restoration:\n%s", strings.Join(accumulatedErrors, "\n"))
	}

	log.Println("Custom nodes restoration process complete.")
	return nil
}