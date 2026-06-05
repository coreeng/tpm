package git

import (
	"errors"
	"testing"
)

func TestValidateRef(t *testing.T) {
	tests := []struct {
		name      string
		ref       string
		mockError error
		wantError bool
	}{
		{
			name:      "valid ref",
			ref:       "main",
			mockError: nil,
			wantError: false,
		},
		{
			name:      "invalid ref",
			ref:       "nonexistent",
			mockError: errors.New("invalid ref"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient()
			mock.ValidateRefFunc = func(ref string) error {
				if ref != tt.ref {
					t.Errorf("Expected ref %q, got %q", tt.ref, ref)
				}
				return tt.mockError
			}

			SetClient(mock)
			defer ResetClient()

			err := ValidateRef(tt.ref)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateRef() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestIsGitRepo(t *testing.T) {
	tests := []struct {
		name   string
		mockIs bool
		want   bool
	}{
		{
			name:   "is git repo",
			mockIs: true,
			want:   true,
		},
		{
			name:   "not git repo",
			mockIs: false,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient()
			mock.IsGitRepoFunc = func() bool {
				return tt.mockIs
			}

			SetClient(mock)
			defer ResetClient()

			got := IsGitRepo()
			if got != tt.want {
				t.Errorf("IsGitRepo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFileContent(t *testing.T) {
	tests := []struct {
		name        string
		ref         string
		filePath    string
		mockContent []byte
		mockError   error
		wantError   bool
	}{
		{
			name:        "successful fetch",
			ref:         "main",
			filePath:    "test.yaml",
			mockContent: []byte("code: test-uuid"),
			mockError:   nil,
			wantError:   false,
		},
		{
			name:        "file not found",
			ref:         "main",
			filePath:    "nonexistent.yaml",
			mockContent: nil,
			mockError:   errors.New("file not found"),
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient()
			mock.GetFileContentFunc = func(ref, filePath string) ([]byte, error) {
				if ref != tt.ref {
					t.Errorf("Expected ref %q, got %q", tt.ref, ref)
				}
				if filePath != tt.filePath {
					t.Errorf("Expected filePath %q, got %q", tt.filePath, filePath)
				}
				return tt.mockContent, tt.mockError
			}

			SetClient(mock)
			defer ResetClient()

			content, err := GetFileContent(tt.ref, tt.filePath)
			if (err != nil) != tt.wantError {
				t.Errorf("GetFileContent() error = %v, wantError %v", err, tt.wantError)
			}
			if !tt.wantError && string(content) != string(tt.mockContent) {
				t.Errorf("GetFileContent() = %q, want %q", content, tt.mockContent)
			}
		})
	}
}

func TestListFiles(t *testing.T) {
	tests := []struct {
		name      string
		ref       string
		pattern   string
		mockFiles []string
		mockError error
		wantError bool
	}{
		{
			name:      "successful list",
			ref:       "main",
			pattern:   "*.yaml",
			mockFiles: []string{"module.yaml", "chapter.yaml"},
			mockError: nil,
			wantError: false,
		},
		{
			name:      "no matches",
			ref:       "main",
			pattern:   "*.txt",
			mockFiles: []string{},
			mockError: nil,
			wantError: false,
		},
		{
			name:      "error listing",
			ref:       "invalid-ref",
			pattern:   "*",
			mockFiles: nil,
			mockError: errors.New("invalid ref"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient()
			mock.ListFilesFunc = func(ref, pattern string) ([]string, error) {
				if ref != tt.ref {
					t.Errorf("Expected ref %q, got %q", tt.ref, ref)
				}
				if pattern != tt.pattern {
					t.Errorf("Expected pattern %q, got %q", tt.pattern, pattern)
				}
				return tt.mockFiles, tt.mockError
			}

			SetClient(mock)
			defer ResetClient()

			files, err := ListFiles(tt.ref, tt.pattern)
			if (err != nil) != tt.wantError {
				t.Errorf("ListFiles() error = %v, wantError %v", err, tt.wantError)
			}
			if !tt.wantError {
				if len(files) != len(tt.mockFiles) {
					t.Errorf("ListFiles() returned %d files, want %d", len(files), len(tt.mockFiles))
				}
			}
		})
	}
}

func TestFindModulesAtRef(t *testing.T) {
	tests := []struct {
		name        string
		ref         string
		mockModules []string
		mockError   error
		wantError   bool
	}{
		{
			name:        "successful find",
			ref:         "main",
			mockModules: []string{"module-a", "module-b"},
			mockError:   nil,
			wantError:   false,
		},
		{
			name:        "no modules",
			ref:         "main",
			mockModules: []string{},
			mockError:   nil,
			wantError:   false,
		},
		{
			name:        "error finding",
			ref:         "invalid-ref",
			mockModules: nil,
			mockError:   errors.New("invalid ref"),
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient()
			mock.FindModulesAtRefFunc = func(ref string) ([]string, error) {
				if ref != tt.ref {
					t.Errorf("Expected ref %q, got %q", tt.ref, ref)
				}
				return tt.mockModules, tt.mockError
			}

			SetClient(mock)
			defer ResetClient()

			modules, err := FindModulesAtRef(tt.ref)
			if (err != nil) != tt.wantError {
				t.Errorf("FindModulesAtRef() error = %v, wantError %v", err, tt.wantError)
			}
			if !tt.wantError {
				if len(modules) != len(tt.mockModules) {
					t.Errorf("FindModulesAtRef() returned %d modules, want %d", len(modules), len(tt.mockModules))
				}
			}
		})
	}
}

func TestCollectCodesAtRef(t *testing.T) {
	tests := []struct {
		name      string
		ref       string
		mockCodes map[string]CodeInfo
		mockError error
		wantError bool
	}{
		{
			name: "successful collection",
			ref:  "main",
			mockCodes: map[string]CodeInfo{
				"uuid-1": {Code: "uuid-1", EntityType: "module", FilePath: "module.yaml", ModuleName: "test-module"},
				"uuid-2": {Code: "uuid-2", EntityType: "chapter", FilePath: "chapter.yaml", ModuleName: "test-module"},
			},
			mockError: nil,
			wantError: false,
		},
		{
			name:      "no codes",
			ref:       "main",
			mockCodes: map[string]CodeInfo{},
			mockError: nil,
			wantError: false,
		},
		{
			name:      "error collecting",
			ref:       "invalid-ref",
			mockCodes: nil,
			mockError: errors.New("invalid ref"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient()
			mock.CollectCodesAtRefFunc = func(ref string) (map[string]CodeInfo, error) {
				if ref != tt.ref {
					t.Errorf("Expected ref %q, got %q", tt.ref, ref)
				}
				return tt.mockCodes, tt.mockError
			}

			SetClient(mock)
			defer ResetClient()

			codes, err := CollectCodesAtRef(tt.ref)
			if (err != nil) != tt.wantError {
				t.Errorf("CollectCodesAtRef() error = %v, wantError %v", err, tt.wantError)
			}
			if !tt.wantError {
				if len(codes) != len(tt.mockCodes) {
					t.Errorf("CollectCodesAtRef() returned %d codes, want %d", len(codes), len(tt.mockCodes))
				}
				for code, info := range tt.mockCodes {
					if gotInfo, exists := codes[code]; !exists {
						t.Errorf("CollectCodesAtRef() missing code %q", code)
					} else if gotInfo.EntityType != info.EntityType {
						t.Errorf("Code %q: EntityType = %q, want %q", code, gotInfo.EntityType, info.EntityType)
					}
				}
			}
		})
	}
}

func TestSetClientAndReset(t *testing.T) {
	// Test that SetClient and ResetClient work correctly
	original := defaultClient

	mock := NewMockClient()
	SetClient(mock)

	if defaultClient != mock {
		t.Error("SetClient did not update defaultClient")
	}

	ResetClient()

	// After reset, should be a new GoGitClient (not the same instance as original)
	if defaultClient == mock {
		t.Error("ResetClient did not reset defaultClient")
	}

	// Verify it's a GoGitClient by type
	if _, ok := defaultClient.(*GoGitClient); !ok {
		t.Errorf("ResetClient did not create GoGitClient, got %T", defaultClient)
	}

	// Restore original for other tests
	defaultClient = original
}
