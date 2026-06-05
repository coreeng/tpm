// Package pathutil provides shared utility functions for path manipulation and file system operations.
//
// This package centralizes path-related functionality that was previously duplicated across multiple packages
// (builder, fixer, module, validator) to reduce code duplication and improve maintainability.
//
// Key functions:
//   - GetRelativeModulePath: Extracts clean relative paths for error messages
//   - FileExists: Checks if a path exists and is a regular file
//   - DirExists: Checks if a path exists and is a directory
//   - GetRepoRoot: Returns the git repository root directory
package pathutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
)

// GetRelativeModulePath extracts the relative module path from a full file path.
// It returns the portion starting with "module/" for cleaner error messages and logs.
//
// Examples:
//   - "/full/path/my-module/module/chapter.yml" → "module/chapter.yml"
//   - "/full/path/module/01-chapter/section.yaml" → "module/01-chapter/section.yaml"
//   - "no-module-in-path/file.yaml" → "file.yaml" (fallback to filename)
func GetRelativeModulePath(fullPath string) string {
	// Find "module/" in the path and return from there
	parts := strings.Split(fullPath, "/module/")
	if len(parts) >= 2 {
		return "module/" + parts[len(parts)-1]
	}
	// Fallback to just the filename if we can't find the pattern
	return filepath.Base(fullPath)
}

// FileExists checks if a path exists and is a regular file (not a directory).
// Returns false if the path doesn't exist, is a directory, or on any error.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// DirExists checks if a path exists and is a directory (not a file).
// Returns false if the path doesn't exist, is a file, or on any error.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetRepoRoot returns the git repository root directory.
// It uses go-git to find the repository root starting from the current working directory.
// This is more robust than simple path manipulation as it properly handles git worktrees and submodules.
//
// Returns an error if:
//   - Unable to determine current directory
//   - Not running from within a git repository
//   - Unable to access git worktree information
func GetRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Use go-git to find the repository root
	repo, err := git.PlainOpenWithOptions(cwd, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return "", fmt.Errorf("not in a git repository: %w", err)
	}

	// Get the worktree to find the root path
	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	return worktree.Filesystem.Root(), nil
}
