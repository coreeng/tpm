package pathutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetRelativeModulePath(t *testing.T) {
	tests := []struct {
		name     string
		fullPath string
		want     string
	}{
		{
			name:     "standard module path",
			fullPath: "/full/path/my-module/module/chapter.yml",
			want:     "module/chapter.yml",
		},
		{
			name:     "nested module path",
			fullPath: "/full/path/module/01-chapter/sections/01-section/section.yaml",
			want:     "module/01-chapter/sections/01-section/section.yaml",
		},
		{
			name:     "module at root",
			fullPath: "/module/file.yaml",
			want:     "module/file.yaml",
		},
		{
			name:     "multiple module occurrences",
			fullPath: "/path/module/ignored/module/actual.yaml",
			want:     "module/actual.yaml",
		},
		{
			name:     "no module in path",
			fullPath: "/full/path/to/file.yaml",
			want:     "file.yaml",
		},
		{
			name:     "relative path with module",
			fullPath: "testdata/my-module/module/module.yaml",
			want:     "module/module.yaml",
		},
		{
			name:     "empty path",
			fullPath: "",
			want:     ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetRelativeModulePath(tt.fullPath)
			if got != tt.want {
				t.Errorf("GetRelativeModulePath(%q) = %q, want %q", tt.fullPath, got, tt.want)
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	// Create a temporary file for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-file.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "existing file",
			path: tmpFile,
			want: true,
		},
		{
			name: "existing directory",
			path: tmpDir,
			want: false, // Should return false for directories
		},
		{
			name: "non-existent path",
			path: filepath.Join(tmpDir, "does-not-exist.txt"),
			want: false,
		},
		{
			name: "empty path",
			path: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FileExists(tt.path)
			if got != tt.want {
				t.Errorf("FileExists(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestDirExists(t *testing.T) {
	// Create a temporary directory and file for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-file.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "existing directory",
			path: tmpDir,
			want: true,
		},
		{
			name: "existing file",
			path: tmpFile,
			want: false, // Should return false for files
		},
		{
			name: "non-existent path",
			path: filepath.Join(tmpDir, "does-not-exist"),
			want: false,
		},
		{
			name: "empty path",
			path: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DirExists(tt.path)
			if got != tt.want {
				t.Errorf("DirExists(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
