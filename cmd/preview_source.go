package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/coreeng/tpm/pkg/pathutil"
)

func previewSourceLabel(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	clean := filepath.Clean(path)
	label := filepath.ToSlash(clean)
	if strings.HasPrefix(label, "module/") {
		return label
	}
	if strings.Contains(label, "/module/") {
		return filepath.ToSlash(pathutil.GetRelativeModulePath(label))
	}
	return label
}

func previewSourceLabelRelative(root, path string) string {
	label := previewSourceLabel(path)
	if label == "" {
		return ""
	}
	if strings.HasPrefix(label, "module/") {
		return label
	}

	root = strings.TrimSpace(root)
	if root == "" {
		return label
	}
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." || rel == "" || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return label
	}
	return filepath.ToSlash(rel)
}

func siblingSource(path, name string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	return filepath.Join(filepath.Dir(path), name)
}

func existingSourceOrFallback(path, fallback string) string {
	if strings.TrimSpace(path) == "" {
		return fallback
	}
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return fallback
}
