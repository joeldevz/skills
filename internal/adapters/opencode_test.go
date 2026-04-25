package adapters

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMergeOpencodeConfig_PreservesUserMCP(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "opencode.json")

	// Installed config (what we just copied)
	installed := map[string]interface{}{
		"mcp": map[string]interface{}{
			"neurox": map[string]interface{}{
				"command": []string{"neurox", "mcp"},
				"enabled": true,
				"type":    "local",
			},
			"context7": map[string]interface{}{
				"command": []string{"context7"},
				"enabled": true,
			},
		},
	}
	data, _ := json.MarshalIndent(installed, "", "  ")
	os.WriteFile(configPath, append(data, '\n'), 0o644)

	// Backup config (user had a custom MCP server)
	backup := map[string]json.RawMessage{}
	userMCP := map[string]interface{}{
		"my-custom-server": map[string]interface{}{
			"command": []string{"my-server"},
			"enabled": true,
		},
		"neurox": map[string]interface{}{
			"command": []string{"old", "command"},
			"enabled": false,
		},
	}
	userMCPJSON, _ := json.Marshal(userMCP)
	backup["mcp"] = userMCPJSON

	err := mergeOpencodeConfig(configPath, backup)
	if err != nil {
		t.Fatalf("mergeOpencodeConfig failed: %v", err)
	}

	// Read result
	result, _ := os.ReadFile(configPath)
	var merged map[string]interface{}
	json.Unmarshal(result, &merged)

	mcp, ok := merged["mcp"].(map[string]interface{})
	if !ok {
		t.Fatal("mcp field missing or wrong type after merge")
	}

	// User's custom server should be preserved
	if _, ok := mcp["my-custom-server"]; !ok {
		t.Error("user's custom MCP server was lost after merge")
	}

	// Neurox should be forced to correct shape (installed wins)
	neurox, ok := mcp["neurox"].(map[string]interface{})
	if !ok {
		t.Fatal("neurox MCP entry missing after merge")
	}

	// neurox.enabled should be true (installed version wins)
	if enabled, _ := neurox["enabled"].(bool); !enabled {
		t.Error("neurox.enabled should be true (installed config wins)")
	}
}

func TestMergeOpencodeConfig_NilBackup(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "opencode.json")

	installed := map[string]interface{}{
		"mcp": map[string]interface{}{
			"neurox": map[string]interface{}{"enabled": true},
		},
	}
	data, _ := json.MarshalIndent(installed, "", "  ")
	os.WriteFile(configPath, append(data, '\n'), 0o644)

	// Should not panic with nil backup
	err := mergeOpencodeConfig(configPath, nil)
	if err != nil {
		t.Fatalf("mergeOpencodeConfig with nil backup failed: %v", err)
	}
}

func TestMergeOpencodeConfig_ForceNeuroxEntry(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "opencode.json")

	// Config without neurox MCP
	installed := map[string]interface{}{
		"mcp": map[string]interface{}{},
	}
	data, _ := json.MarshalIndent(installed, "", "  ")
	os.WriteFile(configPath, append(data, '\n'), 0o644)

	err := mergeOpencodeConfig(configPath, nil)
	if err != nil {
		t.Fatalf("mergeOpencodeConfig failed: %v", err)
	}

	result, _ := os.ReadFile(configPath)
	var merged map[string]interface{}
	json.Unmarshal(result, &merged)

	mcp := merged["mcp"].(map[string]interface{})
	if _, ok := mcp["neurox"]; !ok {
		t.Error("neurox entry should be forced even when not in installed config")
	}
}
