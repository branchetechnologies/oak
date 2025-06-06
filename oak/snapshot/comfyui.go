package snapshot

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"yourproject/comfyui-compactor/oak"
	"yourproject/comfyui-compactor/oak/functions"
	"yourproject/comfyui-compactor/oak/system"
)

// ComfyUISnapshotter orchestrates the creation of a ComfyUI snapshot.
type ComfyUISnapshotter struct {
	ComfyUIPath string
}

// NewComfyUISnapshotter creates a new ComfyUISnapshotter.
func NewComfyUISnapshotter(comfyUIPath string) *ComfyUISnapshotter {
	return &ComfyUISnapshotter{ComfyUIPath: comfyUIPath}
}

// Snapshot creates a comprehensive snapshot of the ComfyUI installation.
func (cs *ComfyUISnapshotter) Snapshot() (*oak.Snapshot, error) {
	snapshot := &oak.Snapshot{
		Timestamp:      time.Now().Format(time.RFC3339),
		OakVersion:     "0.1.0", // TODO: Consider making this dynamic or a constant
		ProgramVersion: runtime.Version(),
		Nodes:          make(map[string]functions.NodeInfo),
		Models:         make(map[string]functions.AssetInfo),
	}

	var err error

	snapshot.Environment, _ = system.GetEnvironmentSnapshot() // Handles its own errors by returning "N/A"
	snapshot.ComfyUI, _ = functions.GetGitInfo(cs.ComfyUIPath) // Handles its own errors by returning "unknown"

	// Get pip freeze requirements (Global Python environment)
	pipFreezeOutput, pipErr := functions.RunCommand("", "pip", "freeze")
	if pipErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: 'pip freeze' failed: %v. Requirements will be empty.\n", pipErr)
		snapshot.Requirements = []string{}
	} else {
		lines := strings.Split(pipFreezeOutput, "\n")
		var reqs []string
		for _, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "#") {
				reqs = append(reqs, trimmedLine)
			}
		}
		sort.Strings(reqs)
		snapshot.Requirements = reqs
	}

	// 4. Scan Custom Nodes
	customNodesBasePath := filepath.Join(cs.ComfyUIPath, "custom_nodes")
	if _, statErr := os.Stat(customNodesBasePath); !os.IsNotExist(statErr) {
		nodeEntries, readDirErr := os.ReadDir(customNodesBasePath)
		if readDirErr != nil {
			return nil, fmt.Errorf("failed to read custom_nodes directory %s: %w", customNodesBasePath, readDirErr)
		}
		for _, entry := range nodeEntries {
			if !entry.IsDir() {
				continue
			}
			nodeDirName := entry.Name()
			nodeDirPath := filepath.Join(customNodesBasePath, nodeDirName)

			nodeSnapper := NewNodeSnapshotter(nodeDirPath, customNodesBasePath)
			nodeInfo, nodeErr := nodeSnapper.Snapshot()
			if nodeErr != nil {
				// Log or handle error for individual node, perhaps continue
				fmt.Fprintf(os.Stderr, "Warning: failed to snapshot node %s: %v\n", nodeDirName, nodeErr)
				continue
			}

			nodeKey := functions.ParseOwnerRepoFromURL(nodeInfo.GitURL)
			if nodeKey == "" {
				nodeKey = nodeInfo.Path // Fallback to relative path
			}
			snapshot.Nodes[nodeKey] = nodeInfo
		}
	}

	// 5. Scan Assets (Models, Loras, etc.)
	// Define asset types and their extensions. This could be made more configurable.
	assetConfigs := []struct {
		AssetType  string
		Extensions []string
	}{
		{"models", []string{".safetensors", ".ckpt", ".json", ".pth", ".bin"}}, // .json for VAEs, .pth/.bin for others
		{"loras", []string{".safetensors", ".ckpt", ".pt"}},
		// Add other asset types like "controlnet", "vae", "embeddings", etc.
	}

	for _, assetConfig := range assetConfigs {
		assetTypeBasePath := filepath.Join(cs.ComfyUIPath, assetConfig.AssetType)
		if _, statErr := os.Stat(assetTypeBasePath); os.IsNotExist(statErr) {
			continue // Skip if asset type directory doesn't exist
		}

		walkErr := filepath.WalkDir(assetTypeBasePath, func(path string, d os.DirEntry, walkErrIn error) error {
			if walkErrIn != nil {
				fmt.Fprintf(os.Stderr, "Warning: error accessing path %s during asset scan: %v\n", path, walkErrIn)
				return nil // Continue walking if possible
			}
			if d.IsDir() {
				return nil // Skip directories
			}

			for _, ext := range assetConfig.Extensions {
				if strings.HasSuffix(strings.ToLower(d.Name()), strings.ToLower(ext)) {
					fileHashVal, hashErr := functions.HashFile(path)
					if hashErr != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to hash asset file %s: %v\n", path, hashErr)
						return nil // Continue to next file
					}

					if _, exists := snapshot.Models[fileHashVal]; !exists {
						assetSnapper := NewAssetSnapshotter(path, assetTypeBasePath)
						assetInfo, assetErr := assetSnapper.Snapshot()
						if assetErr != nil {
							fmt.Fprintf(os.Stderr, "Warning: failed to snapshot asset %s: %v\n", d.Name(), assetErr)
							return nil // Continue to next file
						}
						snapshot.Models[fileHashVal] = assetInfo
					}
					break // Found a matching extension
				}
			}
			return nil
		})

		if walkErr != nil {
			// This error is from filepath.WalkDir itself, not the callback.
			// It might be serious enough to halt, or just log and continue with other asset types.
			fmt.Fprintf(os.Stderr, "Warning: error walking asset directory %s: %v\n", assetTypeBasePath, walkErr)
		}
	}

	return snapshot, err
}