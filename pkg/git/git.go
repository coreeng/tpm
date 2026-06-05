package git

// CodeInfo represents a code found in a git ref
type CodeInfo struct {
	Code       string
	EntityType string
	FilePath   string
	ModuleName string
	ParentCode string // The code of the parent entity (empty for module-level entities)
	ParentType string // The type of the parent entity (empty for module-level entities)
}

// defaultClient is the default GitClient implementation
// Can be overridden for testing using SetClient()
var defaultClient GitClient

func init() {
	// Initialize with GoGitClient - panics if not in a git repo
	client, err := NewGoGitClient()
	if err != nil {
		// If we can't open a git repo, default to nil
		// Commands that need git will fail with appropriate error
		defaultClient = nil
	} else {
		defaultClient = client
	}
}

// SetClient sets the GitClient implementation to use
// This is primarily for testing - allows injecting a mock client
func SetClient(client GitClient) {
	defaultClient = client
}

// ResetClient resets the GitClient to the default GoGitClient
// Useful for cleanup in tests
func ResetClient() {
	client, err := NewGoGitClient()
	if err != nil {
		defaultClient = nil
	} else {
		defaultClient = client
	}
}

// ValidateRef checks if a git ref exists
func ValidateRef(ref string) error {
	return defaultClient.ValidateRef(ref)
}

// IsGitRepo checks if the current directory is inside a git repository
func IsGitRepo() bool {
	return defaultClient.IsGitRepo()
}

// GetFileContent reads a file from a specific git ref without checking it out
func GetFileContent(ref, filePath string) ([]byte, error) {
	return defaultClient.GetFileContent(ref, filePath)
}

// ListFiles lists files matching a pattern in a git ref
func ListFiles(ref, pattern string) ([]string, error) {
	return defaultClient.ListFiles(ref, pattern)
}

// FindModulesAtRef finds all modules at a specific git ref
func FindModulesAtRef(ref string) ([]string, error) {
	return defaultClient.FindModulesAtRef(ref)
}

// CollectCodesAtRef collects all codes from all modules at a specific git ref
func CollectCodesAtRef(ref string) (map[string]CodeInfo, error) {
	return defaultClient.CollectCodesAtRef(ref)
}
