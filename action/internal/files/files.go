package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SafePath resolves a path relative to the workspace root and ensures
// it doesn't escape outside the workspace (prevents path traversal).
func SafePath(workspace, path string) (string, error) {
	abs := path
	if !filepath.IsAbs(path) {
		abs = filepath.Join(workspace, path)
	}
	abs = filepath.Clean(abs)

	if !strings.HasPrefix(abs, filepath.Clean(workspace)+string(filepath.Separator)) && abs != filepath.Clean(workspace) {
		return "", fmt.Errorf("path %q is outside the workspace", path)
	}
	return abs, nil
}

// ReadFile reads a file and returns its contents with line numbers.
func ReadFile(workspace, path string) (string, error) {
	absPath, err := SafePath(workspace, path)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", path, err)
	}

	lines := strings.Split(string(data), "\n")
	var b strings.Builder
	for i, line := range lines {
		fmt.Fprintf(&b, "%4d | %s\n", i+1, line)
	}
	return b.String(), nil
}

// WriteFile creates or overwrites a file with the given content.
// Creates parent directories as needed.
func WriteFile(workspace, path, content string) error {
	absPath, err := SafePath(workspace, path)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return fmt.Errorf("failed to create directories for %s: %w", path, err)
	}

	if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}
	return nil
}

// EditFile performs a find-and-replace in a file.
func EditFile(workspace, path, oldStr, newStr string) error {
	absPath, err := SafePath(workspace, path)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	content := string(data)
	count := strings.Count(content, oldStr)
	if count == 0 {
		return fmt.Errorf("old_string not found in %s", path)
	}
	if count > 1 {
		return fmt.Errorf("old_string found %d times in %s — must be unique", count, path)
	}

	newContent := strings.Replace(content, oldStr, newStr, 1)
	if err := os.WriteFile(absPath, []byte(newContent), 0o644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}
	return nil
}

// ListDirectory lists files and subdirectories at the given path.
func ListDirectory(workspace, path string) (string, error) {
	absPath, err := SafePath(workspace, path)
	if err != nil {
		return "", err
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to list %s: %w", path, err)
	}

	var b strings.Builder
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		fmt.Fprintln(&b, name)
	}
	return b.String(), nil
}

// SearchCode searches for a pattern in files under the given path.
// Returns matching lines with file paths and line numbers.
func SearchCode(workspace, pattern, searchPath string) (string, error) {
	absPath, err := SafePath(workspace, searchPath)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	matchCount := 0
	maxMatches := 100

	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if info.IsDir() {
			// Skip hidden dirs and common non-source dirs
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" || name == "__pycache__" {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Size() > 1024*1024 { // skip files > 1MB
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(workspace, path)
		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			if strings.Contains(line, pattern) {
				fmt.Fprintf(&b, "%s:%d: %s\n", relPath, i+1, line)
				matchCount++
				if matchCount >= maxMatches {
					fmt.Fprintf(&b, "\n... truncated at %d matches\n", maxMatches)
					return fmt.Errorf("max matches reached")
				}
			}
		}
		return nil
	})

	if b.Len() == 0 {
		return "No matches found.", nil
	}
	return b.String(), nil
}
