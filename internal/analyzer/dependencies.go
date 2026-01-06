package analyzer

import (
	"encoding/base64"
	"encoding/json"
	"regexp"
	"sort"
	"strings"

	"github.com/agnivo988/Repo-lyzer/internal/github"
)

// Dependency represents a single dependency
type Dependency struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Type    string `json:"type"` // "production", "dev", "peer", etc.
}

// DependencyFile represents dependencies from a single file
type DependencyFile struct {
	Filename     string       `json:"filename"`
	FileType     string       `json:"file_type"` // "npm", "go", "python", "rust", etc.
	Dependencies []Dependency `json:"dependencies"`
	TotalCount   int          `json:"total_count"`
}

// DependencyAnalysis holds all dependency information for a repo
type DependencyAnalysis struct {
	Files        []DependencyFile `json:"files"`
	TotalDeps    int              `json:"total_deps"`
	Languages    []string         `json:"languages"`
	HasLockFile  bool             `json:"has_lock_file"`
}

// AnalyzeDependencies fetches and parses dependency files from a repository
func AnalyzeDependencies(client *github.Client, owner, repo, branch string, fileTree []github.TreeEntry) (*DependencyAnalysis, error) {
	analysis := &DependencyAnalysis{
		Files:     []DependencyFile{},
		Languages: []string{},
	}

	// Find dependency files in the tree
	depFiles := findDependencyFiles(fileTree)
	
	for _, df := range depFiles {
		content, err := client.GetFileContent(owner, repo, df.path)
		if err != nil {
			continue // Skip files we can't read
		}

		// Decode base64 content
		decoded, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			continue
		}

		var deps []Dependency
		var fileType string

		switch df.fileType {
		case "npm":
			deps, fileType = parsePackageJSON(decoded)
		case "go":
			deps, fileType = parseGoMod(decoded)
		case "python":
			deps, fileType = parseRequirementsTxt(decoded)
		case "rust":
			deps, fileType = parseCargoToml(decoded)
		case "ruby":
			deps, fileType = parseGemfile(decoded)
		}

		if len(deps) > 0 {
			analysis.Files = append(analysis.Files, DependencyFile{
				Filename:     df.path,
				FileType:     fileType,
				Dependencies: deps,
				TotalCount:   len(deps),
			})
			analysis.TotalDeps += len(deps)
			
			// Track language
			if !contains(analysis.Languages, fileType) {
				analysis.Languages = append(analysis.Languages, fileType)
			}
		}
	}

	// Check for lock files
	analysis.HasLockFile = hasLockFile(fileTree)

	return analysis, nil
}

type depFileInfo struct {
	path     string
	fileType string
}

func findDependencyFiles(tree []github.TreeEntry) []depFileInfo {
	var files []depFileInfo
	
	depFilePatterns := map[string]string{
		"package.json":     "npm",
		"go.mod":           "go",
		"requirements.txt": "python",
		"Pipfile":          "python",
		"pyproject.toml":   "python",
		"Cargo.toml":       "rust",
		"Gemfile":          "ruby",
	}

	for _, entry := range tree {
		if entry.Type != "blob" {
			continue
		}
		
		// Get filename from path
		parts := strings.Split(entry.Path, "/")
		filename := parts[len(parts)-1]
		
		if fileType, ok := depFilePatterns[filename]; ok {
			files = append(files, depFileInfo{
				path:     entry.Path,
				fileType: fileType,
			})
		}
	}

	return files
}

func hasLockFile(tree []github.TreeEntry) bool {
	lockFiles := []string{
		"package-lock.json",
		"yarn.lock",
		"pnpm-lock.yaml",
		"go.sum",
		"Pipfile.lock",
		"poetry.lock",
		"Cargo.lock",
		"Gemfile.lock",
	}

	for _, entry := range tree {
		parts := strings.Split(entry.Path, "/")
		filename := parts[len(parts)-1]
		
		for _, lockFile := range lockFiles {
			if filename == lockFile {
				return true
			}
		}
	}
	return false
}

// parsePackageJSON parses npm package.json
func parsePackageJSON(content []byte) ([]Dependency, string) {
	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
		PeerDeps        map[string]string `json:"peerDependencies"`
	}

	if err := json.Unmarshal(content, &pkg); err != nil {
		return nil, "npm"
	}

	var deps []Dependency

	for name, version := range pkg.Dependencies {
		deps = append(deps, Dependency{
			Name:    name,
			Version: cleanVersion(version),
			Type:    "production",
		})
	}

	for name, version := range pkg.DevDependencies {
		deps = append(deps, Dependency{
			Name:    name,
			Version: cleanVersion(version),
			Type:    "dev",
		})
	}

	for name, version := range pkg.PeerDeps {
		deps = append(deps, Dependency{
			Name:    name,
			Version: cleanVersion(version),
			Type:    "peer",
		})
	}

	// Sort by name
	sort.Slice(deps, func(i, j int) bool {
		return deps[i].Name < deps[j].Name
	})

	return deps, "npm"
}

// parseGoMod parses Go go.mod file
func parseGoMod(content []byte) ([]Dependency, string) {
	var deps []Dependency
	lines := strings.Split(string(content), "\n")
	
	inRequire := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "require (") {
			inRequire = true
			continue
		}
		if line == ")" {
			inRequire = false
			continue
		}
		
		// Single line require
		if strings.HasPrefix(line, "require ") && !strings.Contains(line, "(") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				deps = append(deps, Dependency{
					Name:    parts[1],
					Version: parts[2],
					Type:    "production",
				})
			}
			continue
		}
		
		// Inside require block
		if inRequire && line != "" && !strings.HasPrefix(line, "//") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				depType := "production"
				if strings.Contains(line, "// indirect") {
					depType = "indirect"
				}
				deps = append(deps, Dependency{
					Name:    parts[0],
					Version: parts[1],
					Type:    depType,
				})
			}
		}
	}

	return deps, "go"
}

// parseRequirementsTxt parses Python requirements.txt
func parseRequirementsTxt(content []byte) ([]Dependency, string) {
	var deps []Dependency
	lines := strings.Split(string(content), "\n")
	
	// Regex patterns for different requirement formats
	versionPattern := regexp.MustCompile(`^([a-zA-Z0-9_-]+)\s*([=<>!~]+.*)$`)
	simplePattern := regexp.MustCompile(`^([a-zA-Z0-9_-]+)\s*$`)
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
			continue
		}
		
		// Try versioned pattern first
		if matches := versionPattern.FindStringSubmatch(line); len(matches) >= 3 {
			deps = append(deps, Dependency{
				Name:    matches[1],
				Version: strings.TrimSpace(matches[2]),
				Type:    "production",
			})
			continue
		}
		
		// Try simple pattern (just package name)
		if matches := simplePattern.FindStringSubmatch(line); len(matches) >= 2 {
			deps = append(deps, Dependency{
				Name:    matches[1],
				Version: "*",
				Type:    "production",
			})
		}
	}

	return deps, "python"
}

// parseCargoToml parses Rust Cargo.toml
func parseCargoToml(content []byte) ([]Dependency, string) {
	var deps []Dependency
	lines := strings.Split(string(content), "\n")
	
	inDeps := false
	inDevDeps := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if line == "[dependencies]" {
			inDeps = true
			inDevDeps = false
			continue
		}
		if line == "[dev-dependencies]" {
			inDeps = false
			inDevDeps = true
			continue
		}
		if strings.HasPrefix(line, "[") {
			inDeps = false
			inDevDeps = false
			continue
		}
		
		if (inDeps || inDevDeps) && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[0])
				version := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
				
				depType := "production"
				if inDevDeps {
					depType = "dev"
				}
				
				deps = append(deps, Dependency{
					Name:    name,
					Version: version,
					Type:    depType,
				})
			}
		}
	}

	return deps, "rust"
}

// parseGemfile parses Ruby Gemfile
func parseGemfile(content []byte) ([]Dependency, string) {
	var deps []Dependency
	lines := strings.Split(string(content), "\n")
	
	gemPattern := regexp.MustCompile(`gem\s+['"]([^'"]+)['"](?:\s*,\s*['"]([^'"]+)['"])?`)
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		
		if matches := gemPattern.FindStringSubmatch(line); len(matches) >= 2 {
			version := "*"
			if len(matches) >= 3 && matches[2] != "" {
				version = matches[2]
			}
			
			deps = append(deps, Dependency{
				Name:    matches[1],
				Version: version,
				Type:    "production",
			})
		}
	}

	return deps, "ruby"
}

func cleanVersion(v string) string {
	// Remove common prefixes like ^, ~, >=, etc. for display
	v = strings.TrimSpace(v)
	return v
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
