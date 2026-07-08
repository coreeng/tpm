package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/coreeng/tpm/pkg/pathutil"
)

type previewSourceRef struct {
	File     string `json:"file"`
	Property string `json:"property,omitempty"`
}

type previewText struct {
	Value  string            `json:"value"`
	Source *previewSourceRef `json:"source,omitempty"`
}

type previewNumber struct {
	Value  int               `json:"value"`
	Source *previewSourceRef `json:"source,omitempty"`
}

func sourcedText(value, path, property string) previewText {
	return previewText{Value: value, Source: previewSourceRefFor(path, property)}
}

func sourcedNumber(value int, path, property string) previewNumber {
	return previewNumber{Value: value, Source: previewSourceRefFor(path, property)}
}

func previewSourceRefFor(path, property string) *previewSourceRef {
	label := previewSourceLabel(path)
	if label == "" {
		return nil
	}
	property = strings.TrimSpace(property)
	lowerLabel := strings.ToLower(label)
	if !strings.HasSuffix(lowerLabel, ".yaml") && !strings.HasSuffix(lowerLabel, ".yml") {
		property = ""
	}
	return &previewSourceRef{File: label, Property: property}
}

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
