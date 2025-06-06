package functions

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// GitInfo holds information about a Git repository.
type GitInfo struct {
	CommitSHA string `json:"commitsha"`
	GitURL    string `json:"git_url"`
}

// NodeInfo holds information about a custom node.
type NodeInfo struct {
	CommitSHA    string   `json:"commitsha"`
	GitURL       string   `json:"git_url"`
	Path         string   `json:"path"`
	Requirements []string `json:"requirements"`
}

// AssetInfo holds information about an asset (e.g., model).
type AssetInfo struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

// RunCommand executes a shell command and returns its output or an error.
func RunCommand(dir string, commandName string, args ...string) (string, error) {
	cmd := exec.Command(commandName, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("command `%s %s` failed in dir `%s`: %w, stderr: %s", commandName, strings.Join(args, " "), dir, err, stderr.String())
	}
	return strings.TrimSpace(out.String()), nil
}

// GetGitInfo retrieves Git commit SHA and remote URL for a given repository path.
func GetGitInfo(repoPath string) (GitInfo, error) {
	info := GitInfo{CommitSHA: "unknown", GitURL: "unknown"}
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return info, nil // Not a git repo or .git dir not found, return default info
	}

	commitSHA, err := RunCommand(repoPath, "git", "rev-parse", "HEAD")
	if err == nil {
		info.CommitSHA = commitSHA
	} // If error, "unknown" remains

	gitURL, err := RunCommand(repoPath, "git", "config", "--get", "remote.origin.url")
	if err == nil {
		info.GitURL = gitURL
	} // If error, "unknown" remains

	return info, nil
}

// ParseOwnerRepoFromURL extracts "owner/repo" from common Git URL formats.
func ParseOwnerRepoFromURL(gitURL string) string {
	if gitURL == "" || gitURL == "unknown" {
		return ""
	}
	// HTTPS: https://github.com/owner/repo.git or https://gitlab.com/owner/repo
	httpsRegex := regexp.MustCompile(`https?://(?:www\.)?(?:github\.com|gitlab\.com|bitbucket\.org)/([^/]+)/([^/.]+?)(?:\.git)?/?$`)
	matches := httpsRegex.FindStringSubmatch(gitURL)
	if len(matches) == 3 {
		return fmt.Sprintf("%s/%s", matches[1], matches[2])
	}

	// SSH: git@github.com:owner/repo.git or git@gitlab.com:owner/repo
	sshRegex := regexp.MustCompile(`git@(?:github\.com|gitlab\.com|bitbucket\.org):([^/]+)/([^/.]+?)(?:\.git)?/?$`)
	matches = sshRegex.FindStringSubmatch(gitURL)
	if len(matches) == 3 {
		return fmt.Sprintf("%s/%s", matches[1], matches[2])
	}
	return "" // Return empty if no match
}

// GetRequirements reads and returns a sorted list of requirements from a requirements.txt file.
func GetRequirements(repoPath string) ([]string, error) {
	reqPath := filepath.Join(repoPath, "requirements.txt")
	file, err := os.Open(reqPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil // No requirements.txt, return empty list
		}
		return nil, fmt.Errorf("failed to open requirements file %s: %w", reqPath, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") { // Ignore empty lines and comments
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading requirements file %s: %w", reqPath, err)
	}
	sort.Strings(lines)
	return lines, nil
}

// HashFile computes the SHA256 hash of a file.
func HashFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s for hashing: %w", filePath, err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to hash file %s: %w", filePath, err)
	}
	return fmt.Sprintf("sha256:%x", hash.Sum(nil)), nil
}

// ScanCustomNodes scans the custom_nodes directory for ComfyUI and gathers info.
func ScanCustomNodes(comfyUIPath string) (map[string]NodeInfo, error) {
	nodes := make(map[string]NodeInfo)
	customNodesPath := filepath.Join(comfyUIPath, "custom_nodes")

	if _, err := os.Stat(customNodesPath); os.IsNotExist(err) {
		return nodes, nil // custom_nodes directory doesn't exist
	}

	entries, err := os.ReadDir(customNodesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read custom_nodes directory %s: %w", customNodesPath, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		nodeDirName := entry.Name()
		nodePath := filepath.Join(customNodesPath, nodeDirName)
		// Use forward slashes for relative paths in output, consistent with Python version
		relPath := filepath.ToSlash(filepath.Join("custom_nodes", nodeDirName))

		var gitInfoData GitInfo
		var reqs []string

		// Check if .git directory exists to determine if it's a git repo
		gitRepoIndicatorPath := filepath.Join(nodePath, ".git")
		if _, statErr := os.Stat(gitRepoIndicatorPath); !os.IsNotExist(statErr) {
			gitInfoData, _ = GetGitInfo(nodePath) // Errors are handled in GetGitInfo by returning "unknown"
		} else {
			gitInfoData = GitInfo{CommitSHA: "unknown", GitURL: "unknown"}
		}

		reqs, _ = GetRequirements(nodePath) // Errors are handled in GetRequirements by returning empty slice or logging

		ownerRepo := ParseOwnerRepoFromURL(gitInfoData.GitURL)
		nodeKey := ownerRepo
		if nodeKey == "" { // If owner/repo couldn't be parsed (e.g., not a git repo or unknown URL format)
			nodeKey = relPath // Fallback to relative path as key
		}

		// Ensure no duplicate keys, though ownerRepo should be unique if present
		if _, exists := nodes[nodeKey]; !exists {
			nodes[nodeKey] = NodeInfo{
				CommitSHA:    gitInfoData.CommitSHA,
				GitURL:       gitInfoData.GitURL,
				Path:         relPath,
				Requirements: reqs,
			}
		}
	}
	return nodes, nil
}

// ScanAssets scans specified asset directories for files with given extensions.
func ScanAssets(comfyUIPath string, assetTypes []string, extensions []string) (map[string]AssetInfo, error) {
	scannedAssets := make(map[string]AssetInfo)

	for _, assetType := range assetTypes {
		baseDir := filepath.Join(comfyUIPath, assetType)
		if _, err := os.Stat(baseDir); os.IsNotExist(err) {
			continue // Skip if asset type directory doesn't exist
		}

		err := filepath.WalkDir(baseDir, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: error accessing path %s: %v\n", path, walkErr)
				return nil // Continue walking if possible
			}
			if d.IsDir() {
				return nil // Skip directories
			}

			for _, ext := range extensions {
				if strings.HasSuffix(strings.ToLower(d.Name()), strings.ToLower(ext)) {
					fileHashVal, hashErr := HashFile(path)
					if hashErr != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to hash file %s: %v\n", path, hashErr)
						return nil // Continue to next file
					}

					if _, exists := scannedAssets[fileHashVal]; !exists {
						// Calculate relative path from the assetType base directory (e.g., "models")
						// to the directory containing the asset file.
						relPathToDir, relErr := filepath.Rel(baseDir, filepath.Dir(path))
						if relErr != nil {
							fmt.Fprintf(os.Stderr, "Warning: failed to get relative path for %s: %v\n", path, relErr)
							relPathToDir = "." // Default to current dir if error
						}
						
						outputPathStr := ""
						if relPathToDir != "." { // Only add slash if relPathToDir is not "."
							outputPathStr = filepath.ToSlash(relPathToDir) + "/"
						}


						scannedAssets[fileHashVal] = AssetInfo{
							Path: outputPathStr,
							Name: d.Name(),
						}
					}
					break // Found a matching extension, no need to check others for this file
				}
			}
			return nil
		})
		if err != nil {
			// This error is from filepath.WalkDir itself, not the callback.
			return nil, fmt.Errorf("error walking asset directory %s: %w", baseDir, err)
		}
	}
	return scannedAssets, nil
}