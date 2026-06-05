package git

// MockClient is a mock implementation of GitClient for testing
// Each method can be customized by setting the corresponding function field
type MockClient struct {
	ValidateRefFunc       func(ref string) error
	IsGitRepoFunc         func() bool
	GetFileContentFunc    func(ref, filePath string) ([]byte, error)
	ListFilesFunc         func(ref, pattern string) ([]string, error)
	FindModulesAtRefFunc  func(ref string) ([]string, error)
	CollectCodesAtRefFunc func(ref string) (map[string]CodeInfo, error)
}

// NewMockClient creates a new MockClient with default implementations
// that return nil/empty/false values
func NewMockClient() *MockClient {
	return &MockClient{}
}

// ValidateRef calls the ValidateRefFunc if set, otherwise returns nil
func (m *MockClient) ValidateRef(ref string) error {
	if m.ValidateRefFunc != nil {
		return m.ValidateRefFunc(ref)
	}
	return nil
}

// IsGitRepo calls the IsGitRepoFunc if set, otherwise returns true
func (m *MockClient) IsGitRepo() bool {
	if m.IsGitRepoFunc != nil {
		return m.IsGitRepoFunc()
	}
	return true
}

// GetFileContent calls the GetFileContentFunc if set, otherwise returns empty bytes
func (m *MockClient) GetFileContent(ref, filePath string) ([]byte, error) {
	if m.GetFileContentFunc != nil {
		return m.GetFileContentFunc(ref, filePath)
	}
	return []byte{}, nil
}

// ListFiles calls the ListFilesFunc if set, otherwise returns empty slice
func (m *MockClient) ListFiles(ref, pattern string) ([]string, error) {
	if m.ListFilesFunc != nil {
		return m.ListFilesFunc(ref, pattern)
	}
	return []string{}, nil
}

// FindModulesAtRef calls the FindModulesAtRefFunc if set, otherwise returns empty slice
func (m *MockClient) FindModulesAtRef(ref string) ([]string, error) {
	if m.FindModulesAtRefFunc != nil {
		return m.FindModulesAtRefFunc(ref)
	}
	return []string{}, nil
}

// CollectCodesAtRef calls the CollectCodesAtRefFunc if set, otherwise returns empty map
func (m *MockClient) CollectCodesAtRef(ref string) (map[string]CodeInfo, error) {
	if m.CollectCodesAtRefFunc != nil {
		return m.CollectCodesAtRefFunc(ref)
	}
	return make(map[string]CodeInfo), nil
}
