package agent

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/AkshayNayak/ticketflow/action/internal/files"
	"github.com/AkshayNayak/ticketflow/action/internal/provider"
)

// ToolDefinitions returns the provider-agnostic list of tools available to the AI agent.
func ToolDefinitions() []provider.Tool {
	return []provider.Tool{
		{
			Name:        "read_file",
			Description: "Read the contents of a file. Returns the file with line numbers.",
			Parameters: map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The file path relative to the repository root.",
				},
			},
			Required: []string{"path"},
		},
		{
			Name:        "write_file",
			Description: "Create a new file or completely overwrite an existing file with the given content.",
			Parameters: map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The file path relative to the repository root.",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "The complete file content to write.",
				},
			},
			Required: []string{"path", "content"},
		},
		{
			Name:        "edit_file",
			Description: "Edit a file by replacing a specific string with a new string. The old_string must appear exactly once in the file.",
			Parameters: map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The file path relative to the repository root.",
				},
				"old_string": map[string]interface{}{
					"type":        "string",
					"description": "The exact string to find and replace. Must be unique in the file.",
				},
				"new_string": map[string]interface{}{
					"type":        "string",
					"description": "The string to replace old_string with.",
				},
			},
			Required: []string{"path", "old_string", "new_string"},
		},
		{
			Name:        "list_directory",
			Description: "List files and subdirectories at the given path.",
			Parameters: map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The directory path relative to the repository root. Use '.' for the root.",
				},
			},
			Required: []string{"path"},
		},
		{
			Name:        "search_code",
			Description: "Search for a text pattern in files under the given directory. Returns matching lines with file paths and line numbers.",
			Parameters: map[string]interface{}{
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "The text pattern to search for.",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The directory to search in, relative to repo root. Use '.' for the entire repo.",
				},
			},
			Required: []string{"pattern"},
		},
		{
			Name:        "run_command",
			Description: "Execute a shell command in the repository directory. Use for running tests, linters, or build commands. Commands are sandboxed to the repository.",
			Parameters: map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "The shell command to execute.",
				},
			},
			Required: []string{"command"},
		},
	}
}

// HandleToolCall executes a tool call and returns the result string.
func HandleToolCall(workspace string, name string, inputRaw json.RawMessage, tracker *ChangeTracker) (string, bool) {
	var input map[string]interface{}
	if err := json.Unmarshal(inputRaw, &input); err != nil {
		return fmt.Sprintf("Error parsing tool input: %v", err), true
	}

	switch name {
	case "read_file":
		path, _ := input["path"].(string)
		result, err := files.ReadFile(workspace, path)
		if err != nil {
			return err.Error(), true
		}
		return result, false

	case "write_file":
		path, _ := input["path"].(string)
		content, _ := input["content"].(string)
		if err := files.WriteFile(workspace, path, content); err != nil {
			return err.Error(), true
		}
		tracker.Track(path)
		return fmt.Sprintf("Successfully wrote %s", path), false

	case "edit_file":
		path, _ := input["path"].(string)
		oldStr, _ := input["old_string"].(string)
		newStr, _ := input["new_string"].(string)
		if err := files.EditFile(workspace, path, oldStr, newStr); err != nil {
			return err.Error(), true
		}
		tracker.Track(path)
		return fmt.Sprintf("Successfully edited %s", path), false

	case "list_directory":
		path, _ := input["path"].(string)
		if path == "" {
			path = "."
		}
		result, err := files.ListDirectory(workspace, path)
		if err != nil {
			return err.Error(), true
		}
		return result, false

	case "search_code":
		pattern, _ := input["pattern"].(string)
		path, _ := input["path"].(string)
		if path == "" {
			path = "."
		}
		result, err := files.SearchCode(workspace, pattern, path)
		if err != nil {
			return err.Error(), true
		}
		return result, false

	case "run_command":
		command, _ := input["command"].(string)
		result, err := runCommand(workspace, command)
		if err != nil {
			return fmt.Sprintf("Command failed: %v\n%s", err, result), true
		}
		return result, false

	default:
		return fmt.Sprintf("Unknown tool: %s", name), true
	}
}

// runCommand executes a shell command in the workspace directory with a timeout.
func runCommand(workspace, command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = workspace

	done := make(chan error, 1)
	var output []byte

	go func() {
		var err error
		output, err = cmd.CombinedOutput()
		done <- err
	}()

	select {
	case err := <-done:
		result := string(output)
		if len(result) > 50000 {
			result = result[:50000] + "\n... output truncated"
		}
		return result, err
	case <-time.After(120 * time.Second):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return "Command timed out after 120 seconds", fmt.Errorf("timeout")
	}
}

// ChangeTracker keeps track of files created or modified by the agent.
type ChangeTracker struct {
	files map[string]bool
}

// NewChangeTracker creates a new ChangeTracker.
func NewChangeTracker() *ChangeTracker {
	return &ChangeTracker{files: make(map[string]bool)}
}

// Track records a file as changed.
func (ct *ChangeTracker) Track(path string) {
	ct.files[path] = true
}

// Files returns the list of changed file paths.
func (ct *ChangeTracker) Files() []string {
	result := make([]string, 0, len(ct.files))
	for f := range ct.files {
		result = append(result, f)
	}
	return result
}

// FilesString returns the changed files as a comma-separated string.
func (ct *ChangeTracker) FilesString() string {
	return strings.Join(ct.Files(), ",")
}
