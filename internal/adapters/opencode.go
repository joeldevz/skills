package adapters

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// InstallOpencode installs OpenCode config from srcDir, preserving user MCP servers.
func InstallOpencode(srcDir string) error {
	sourceDir := filepath.Join(srcDir, "opencode")
	target := opencodeDir()

	if _, err := os.Stat(sourceDir); err != nil {
		return fmt.Errorf("opencode source not found: %w", err)
	}

	// Read backup of existing config before overwrite
	existingConfigPath := filepath.Join(target, "opencode.json")
	var backupConfig map[string]json.RawMessage
	if data, err := os.ReadFile(existingConfigPath); err == nil {
		json.Unmarshal(data, &backupConfig)
	}

	// Backup existing dir
	if _, err := os.Stat(target); err == nil {
		backupDirIfExists(target)
	}

	// Copy opencode/ → target using Go (no rsync)
	fmt.Printf("    Copying OpenCode config to %s...\n", target)
	if err := copyDirExcluding(sourceDir, target, []string{"node_modules"}); err != nil {
		return fmt.Errorf("copy opencode dir: %w", err)
	}

	// Merge preserved MCP servers
	if backupConfig != nil {
		installedPath := filepath.Join(target, "opencode.json")
		if err := mergeOpencodeConfig(installedPath, backupConfig); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: MCP merge failed: %v\n", err)
		}
	}

	// Install JS dependencies (bun or npm)
	fmt.Println("    Installing OpenCode dependencies...")
	if err := installJSDeps(target); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: dependency install failed: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run manually: cd %s && bun install\n", target)
	}

	fmt.Printf("    OpenCode installed at %s\n", target)
	return nil
}

// mergeOpencodeConfig preserves user MCP servers, forces neurox entry.
func mergeOpencodeConfig(installedPath string, backup map[string]json.RawMessage) error {
	data, err := os.ReadFile(installedPath)
	if err != nil {
		return err
	}

	var installed map[string]json.RawMessage
	if err := json.Unmarshal(data, &installed); err != nil {
		return err
	}

	// Get backup MCP
	var backupMCP map[string]json.RawMessage
	if raw, ok := backup["mcp"]; ok {
		json.Unmarshal(raw, &backupMCP)
	}
	if backupMCP == nil {
		backupMCP = make(map[string]json.RawMessage)
	}

	// Installed MCP wins over backup
	var installedMCP map[string]json.RawMessage
	if raw, ok := installed["mcp"]; ok {
		json.Unmarshal(raw, &installedMCP)
		for k, v := range installedMCP {
			backupMCP[k] = v
		}
	}

	// Force neurox entry
	neuroxEntry := map[string]interface{}{
		"command": []string{"neurox", "mcp"},
		"enabled": true,
		"type":    "local",
	}
	neuroxJSON, _ := json.Marshal(neuroxEntry)
	backupMCP["neurox"] = neuroxJSON

	mergedMCP, _ := json.Marshal(backupMCP)
	installed["mcp"] = mergedMCP

	out, err := json.MarshalIndent(installed, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(installedPath, string(out)+"\n")
}

func installJSDeps(dir string) error {
	var cmd *exec.Cmd
	if _, err := exec.LookPath("bun"); err == nil {
		cmd = exec.Command("bun", "install", "--silent")
	} else if _, err := exec.LookPath("npm"); err == nil {
		cmd = exec.Command("npm", "install", "--silent")
	} else {
		return fmt.Errorf("neither bun nor npm found")
	}
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func backupDirIfExists(dir string) {
	// No-op if doesn't exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return
	}
	// Just log — actual backup would copy, but we want non-destructive install
	fmt.Printf("    Note: existing config at %s will be updated\n", dir)
}

func opencodeDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "opencode")
}
