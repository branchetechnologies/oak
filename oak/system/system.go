package system

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"yourproject/comfyui-compactor/oak/functions" // Import the local functions package
)

// EnvironmentSnapshot holds a snapshot of the system environment.
type EnvironmentSnapshot struct {
	OS       string       `json:"os"`
	CUDA     CudaInfo     `json:"cuda"`
	Instance InstanceInfo `json:"instance"`
}

// CudaInfo holds information about the CUDA installation.
type CudaInfo struct {
	Version       string `json:"version"`
	DriverVersion string `json:"driver_version"`
}

// InstanceInfo holds information about the system instance (CPU, Memory).
type InstanceInfo struct {
	CPU    CPUInfo    `json:"cpu"`
	Memory MemoryInfo `json:"memory"`
}

// CPUInfo holds information about the CPU.
type CPUInfo struct {
	Model string `json:"model"`
	Cores int    `json:"cores"`
}

// MemoryInfo holds information about system memory.
type MemoryInfo struct {
	TotalPhysical string `json:"total_physical"`
}

// GetOSInfo returns the operating system.
func GetOSInfo() string {
	return runtime.GOOS
}

// GetCudaInfo attempts to retrieve CUDA version and driver version using nvidia-smi.
func GetCudaInfo() (CudaInfo, error) {
	info := CudaInfo{Version: "N/A", DriverVersion: "N/A"}
	// Use functions.RunCommand from the imported package
	output, err := functions.RunCommand("", "nvidia-smi", "--query-gpu=driver_version,cuda_version", "--format=csv,noheader,nounits")
	if err != nil {
		// nvidia-smi might not be installed or not in PATH, or no NVIDIA GPU.
		// This is not a fatal error for the snapshot, so return default info.
		return info, nil
	}

	parts := strings.Split(output, ", ")
	if len(parts) == 2 {
		info.DriverVersion = strings.TrimSpace(parts[0])
		info.Version = strings.TrimSpace(parts[1])
	}
	return info, nil
}

// parseInt is a helper function to parse an integer from a string.
func parseInt(s string) (int, error) {
	i64, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return int(i64), nil
}

// GetMemoryInfo attempts to retrieve total physical memory.
func GetMemoryInfo() (string, error) {
	switch runtime.GOOS {
	case "linux":
		file, err := os.Open("/proc/meminfo")
		if err != nil {
			return "N/A", nil // Not fatal, return N/A
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "MemTotal:") {
				parts := strings.Fields(line) // Splits by whitespace
				if len(parts) >= 2 {
					kbValue, err := parseInt(parts[1])
					if err == nil {
						gbValue := float64(kbValue) / (1024 * 1024) // KB to GB
						return fmt.Sprintf("%.1f GB", gbValue), nil
					}
				}
			}
		}
		return "N/A", nil // MemTotal not found or parse error
	case "windows":
		// Use functions.RunCommand from the imported package
		output, err := functions.RunCommand("", "systeminfo")
		if err != nil {
			return "N/A", nil // systeminfo command failed
		}
		// Regex for "Total Physical Memory:       15,284 MB" (English)
		// or "Gesamter physischer Speicher:  15.284 MB" (German example)
		re := regexp.MustCompile(`(?:Total Physical Memory|Gesamter physischer Speicher):\s*([\d.,]+)\s*MB`)
		matches := re.FindStringSubmatch(output)
		if len(matches) >= 2 {
			mbStr := strings.ReplaceAll(matches[1], ",", "") // Remove thousand separators (,)
			mbStr = strings.ReplaceAll(mbStr, ".", "")       // Remove thousand separators (.) for some locales
			mbValue, err := parseInt(mbStr)
			if err == nil {
				gbValue := float64(mbValue) / 1024.0 // MB to GB
				return fmt.Sprintf("%.1f GB", gbValue), nil
			}
		}
		return "N/A", nil // Regex did not match or parse error
	default:
		return "N/A", nil // Unsupported OS for memory info
	}
}

// GetCPUInfo retrieves CPU model (architecture) and number of cores.
func GetCPUInfo() (CPUInfo, error) {
	// runtime.GOARCH gives the architecture (e.g., amd64, arm64)
	// runtime.NumCPU() gives the number of logical CPUs usable by the current process.
	return CPUInfo{
		Model: runtime.GOARCH,
		Cores: runtime.NumCPU(),
	}, nil
}

// GetEnvironmentSnapshot gathers all environment information.
func GetEnvironmentSnapshot() (EnvironmentSnapshot, error) {
	osInfo := GetOSInfo()
	cudaInfo, _ := GetCudaInfo()       // Errors handled in GetCudaInfo, returns N/A
	cpuInfo, _ := GetCPUInfo()         // This function as defined doesn't return an error
	memInfo, _ := GetMemoryInfo()      // Errors handled in GetMemoryInfo, returns N/A

	return EnvironmentSnapshot{
		OS:   osInfo,
		CUDA: cudaInfo,
		Instance: InstanceInfo{
			CPU:    cpuInfo,
			Memory: MemoryInfo{TotalPhysical: memInfo},
		},
	}, nil
}