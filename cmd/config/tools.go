package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// GetToolPath returns the path to an external tool in the PS4 toolchain.
func GetToolPath(name string) (string, error) {
	toolPath := filepath.Join(GetToolsDir(), name)

	// Check if the tool exists and is executable.
	info, err := os.Stat(toolPath)
	if err != nil {
		return "", fmt.Errorf("tool '%s' not found at %s. Please download it and place it there", name, toolPath)
	}
	if info.IsDir() {
		return "", fmt.Errorf("tool '%s' is a directory, not an executable", name)
	}
	if filepath.Ext(name) != ".py" && info.Mode()&0111 == 0 {
		return "", fmt.Errorf("tool '%s' exists but is not executable. Please run 'chmod +x %s'", name, toolPath)
	}

	return toolPath, nil
}

// RunTool executes an external tool with the given arguments.
func RunTool(name string, args ...string) error {
	toolPath, err := GetToolPath(name)
	if err != nil {
		return err
	}

	var cmd *exec.Cmd
	if filepath.Ext(name) == ".py" {
		cmd = exec.Command("python3", append([]string{toolPath}, args...)...)
	} else {
		cmd = exec.Command(toolPath, args...)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("failed to run tool '%s': %v", name, err)
	}

	return nil
}
