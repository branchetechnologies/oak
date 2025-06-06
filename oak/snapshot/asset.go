package snapshot

import (
	"fmt"
	"path/filepath"

	"yourproject/comfyui-compactor/oak/functions"
)

// AssetSnapshotter handles snapshotting a single asset file.
type AssetSnapshotter struct {
	AssetFilePath    string // Full path to the specific asset file
	AssetTypeBaseDir string // Full path to the base directory for this asset's type (e.g., ".../ComfyUI/models")
}

// NewAssetSnapshotter creates a new AssetSnapshotter.
// assetFilePath is the full path to the specific asset file.
// assetTypeBaseDir is the full path to the root directory for this asset's type (e.g., "models", "loras").
func NewAssetSnapshotter(assetFilePath string, assetTypeBaseDir string) *AssetSnapshotter {
	return &AssetSnapshotter{
		AssetFilePath:    assetFilePath,
		AssetTypeBaseDir: assetTypeBaseDir,
	}
}

// Snapshot creates an AssetInfo structure for the specific asset file.
func (as *AssetSnapshotter) Snapshot() (functions.AssetInfo, error) {
	var assetInfo functions.AssetInfo

	assetInfo.Name = filepath.Base(as.AssetFilePath)


	assetDir := filepath.Dir(as.AssetFilePath)
	relPathToDir, err := filepath.Rel(as.AssetTypeBaseDir, assetDir)
	if err != nil {
		return functions.AssetInfo{}, fmt.Errorf("failed to calculate relative path for asset %s (base: %s): %w", as.AssetFilePath, as.AssetTypeBaseDir, err)
	}


	if relPathToDir == "." {
		assetInfo.Path = ""
	} else {
		assetInfo.Path = filepath.ToSlash(relPathToDir) + "/"
	}


	return assetInfo, nil
}
