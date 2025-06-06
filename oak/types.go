package oak

import (
	"yourproject/comfyui-compactor/oak/functions"
	"yourproject/comfyui-compactor/oak/system"
)

// Snapshot is the top-level structure for the ComfyUI snapshot.
type Snapshot struct {
	Timestamp      string                   `json:"timestamp"`
	OakVersion     string                   `json:"oak_version"`      // Version of this snapshot tool
	ProgramVersion string                   `json:"program_version"`  // Go runtime version
	Environment    system.EnvironmentSnapshot `json:"environment"`
	ComfyUI        functions.GitInfo        `json:"comfyui"`
	Nodes          map[string]functions.NodeInfo `json:"nodes"`
	Models         map[string]functions.AssetInfo `json:"models"`
	Requirements   []string                 `json:"requirements"`   // Global Python requirements
}