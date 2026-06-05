package git

// GitClient defines the interface for git operations
// This interface allows for easy testing by enabling mock implementations
type GitClient interface {
	// ValidateRef checks if a git ref exists
	ValidateRef(ref string) error

	// IsGitRepo checks if the current directory is inside a git repository
	IsGitRepo() bool

	// GetFileContent reads a file from a specific git ref without checking it out
	GetFileContent(ref, filePath string) ([]byte, error)

	// ListFiles lists files matching a pattern in a git ref
	ListFiles(ref, pattern string) ([]string, error)

	// FindModulesAtRef finds all modules at a specific git ref
	FindModulesAtRef(ref string) ([]string, error)

	// CollectCodesAtRef collects all codes from all modules at a specific git ref
	CollectCodesAtRef(ref string) (map[string]CodeInfo, error)
}
