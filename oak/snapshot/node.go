package snapshot

import (
	"fmt"
	"path/filepath"

	"yourproject/comfyui-compactor/oak/functions"
)

// NodeSnapshotter handles snapshotting a single custom node.
type NodeSnapshotter struct {
	NodePath            string // Full path to the specific custom node's directory
	CustomNodesBasePath string // Full path to the main 'custom_nodes' directory
}

// NewNodeSnapshotter creates a new NodeSnapshotter.
// nodePath is the full path to the specific custom node's directory.
// customNodesBasePath is the full path to the root 'custom_nodes' directory.
func NewNodeSnapshotter(nodePath string, customNodesBasePath string) *NodeSnapshotter {
	return &NodeSnapshotter{
		NodePath:            nodePath,
		CustomNodesBasePath: customNodesBasePath,
	}
}

// Snapshot creates a NodeInfo structure for the specific custom node.
func (ns *NodeSnapshotter) Snapshot() (functions.NodeInfo, error) {
	var nodeInfo functions.NodeInfo
	var err error

	// 1. Get GitInfo for the node
	// GetGitInfo handles its own errors by returning "unknown" for fields if git commands fail.
	gitData, _ := functions.GetGitInfo(ns.NodePath)
	nodeInfo.CommitSHA = gitData.CommitSHA
	nodeInfo.GitURL = gitData.GitURL

	// 2. Get requirements for the node
	// GetRequirements handles its own errors (e.g., file not found returns empty slice).
	reqs, reqErr := functions.GetRequirements(ns.NodePath)
	if reqErr != nil {
		// This would be an unexpected error, like permission issues.
		return functions.NodeInfo{}, fmt.Errorf("failed to get requirements for node %s: %w", ns.NodePath, reqErr)
	}
	nodeInfo.Requirements = reqs

	// 3. Calculate relative path
	// The path should be relative to the custom_nodes directory and use forward slashes.
	relPath, relErr := filepath.Rel(ns.CustomNodesBasePath, ns.NodePath)
	if relErr != nil {
		// This error should ideally not happen if paths are correct.
		return functions.NodeInfo{}, fmt.Errorf("failed to calculate relative path for node %s (base: %s): %w", ns.NodePath, ns.CustomNodesBasePath, relErr)
	}
	nodeInfo.Path = filepath.ToSlash(relPath)

	return nodeInfo, err
}