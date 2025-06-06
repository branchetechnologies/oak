package snapshot

import (
	"yourproject/comfyui-compactor/oak" // For oak.Snapshot type
	// The detailed logic and its imports are now in comfyui.go, node.go, asset.go
)

// ExecuteSnapshot creates a comprehensive snapshot of the ComfyUI installation
// by using the ComfyUISnapshotter.
func ExecuteSnapshot(comfyUIPath string) (*oak.Snapshot, error) {
	// NewComfyUISnapshotter is defined in comfyui.go within the same package.
	snapshotter := NewComfyUISnapshotter(comfyUIPath)
	return snapshotter.Snapshot()
}