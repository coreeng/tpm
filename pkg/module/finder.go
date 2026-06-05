package module

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/coreeng/tpm/pkg/pathutil"
)

// FindModules scans the root directory and returns names of directories
// that contain training platform modules (those with module/module.yaml)
func FindModules(rootDir string) ([]string, error) {
	var modules []string

	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", rootDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip hidden directories and common non-module directories
		name := entry.Name()
		if name[0] == '.' || name == "node_modules" || name == "vendor" {
			continue
		}

		// Check if this directory contains a module definition
		// Try both .yaml and .yml extensions
		moduleYamlPath := filepath.Join(rootDir, name, "module", "module.yaml")
		moduleYmlPath := filepath.Join(rootDir, name, "module", "module.yml")

		if pathutil.FileExists(moduleYamlPath) || pathutil.FileExists(moduleYmlPath) {
			modules = append(modules, name)
		}
	}

	return modules, nil
}

// GetModulePath returns the full path to the module directory
func GetModulePath(rootDir, moduleName string) string {
	return filepath.Join(rootDir, moduleName, "module")
}

// GetModuleFilePath returns the path to the module.yaml file
// Returns the path even if it doesn't exist (for error reporting)
func GetModuleFilePath(rootDir, moduleName string) string {
	modulePath := GetModulePath(rootDir, moduleName)

	// Try .yaml first, then .yml
	yamlPath := filepath.Join(modulePath, "module.yaml")
	if pathutil.FileExists(yamlPath) {
		return yamlPath
	}

	ymlPath := filepath.Join(modulePath, "module.yml")
	if pathutil.FileExists(ymlPath) {
		return ymlPath
	}

	// Return .yaml as default for error messages
	return yamlPath
}
