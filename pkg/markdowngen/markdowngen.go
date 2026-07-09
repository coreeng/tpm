package markdowngen

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/coreeng/tpm/pkg/module"
	"github.com/coreeng/tpm/pkg/pathutil"
)

// GenerateResult contains the results of a markdown generation operation
type GenerateResult struct {
	FilesCreated int
	Errors       []error
}

// GenerateMissingMarkdown creates placeholder markdown files where they are missing
//
// Behavior:
//   - If markdown file exists: OK (nothing to do)
//   - If markdown file doesn't exist: CREATE placeholder markdown file
//
// Note: Schema validation ensures description/successMessage fields are NOT in YAML,
// so there is never any content to migrate from YAML to markdown.
func GenerateMissingMarkdown(rootDir, moduleName string) (*GenerateResult, error) {
	modulePath := module.GetModulePath(rootDir, moduleName)
	mod, err := module.LoadModule(rootDir, moduleName)
	if err != nil {
		return nil, fmt.Errorf("failed to load module: %w", err)
	}
	return generateMissingMarkdownForModule(mod, modulePath)
}

func GenerateMissingMarkdownPath(modulePath string) (*GenerateResult, error) {
	mod, resolved, err := module.LoadPath(modulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load module: %w", err)
	}
	return generateMissingMarkdownForModule(mod, resolved.SourcePath)
}

func generateMissingMarkdownForModule(mod *module.Module, modulePath string) (*GenerateResult, error) {
	result := &GenerateResult{}

	// Process module description.md
	created, err := ensureMarkdownFile(modulePath, "description.md")
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("module: %v", err))
	} else if created {
		result.FilesCreated++
	}

	// Process chapters
	for _, chapter := range mod.Chapters {
		chapterPath := filepath.Dir(chapter.FilePath)
		chRelPath := pathutil.GetRelativeModulePath(chapter.FilePath)

		// Chapter description.md
		created, err := ensureMarkdownFile(chapterPath, "description.md")
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("%s: %v", chRelPath, err))
		} else if created {
			result.FilesCreated++
		}

		// Process sections
		for _, section := range chapter.Sections {
			sectionPath := filepath.Dir(section.FilePath)
			secRelPath := pathutil.GetRelativeModulePath(section.FilePath)

			created, err := ensureMarkdownFile(sectionPath, "description.md")
			if err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("%s: %v", secRelPath, err))
			} else if created {
				result.FilesCreated++
			}
		}

		// Process assessments
		for _, assessment := range chapter.Assessments {
			assessmentPath := filepath.Dir(assessment.FilePath)
			assessRelPath := pathutil.GetRelativeModulePath(assessment.FilePath)

			created, err := ensureMarkdownFile(assessmentPath, "description.md")
			if err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("%s: %v", assessRelPath, err))
			} else if created {
				result.FilesCreated++
			}

			// Process challenges
			for _, challenge := range assessment.Challenges {
				challengePath := filepath.Dir(challenge.FilePath)
				challengeRelPath := pathutil.GetRelativeModulePath(challenge.FilePath)

				// Challenge description.md
				created, err := ensureMarkdownFile(challengePath, "description.md")
				if err != nil {
					result.Errors = append(result.Errors, fmt.Errorf("%s: %v", challengeRelPath, err))
				} else if created {
					result.FilesCreated++
				}

				// Challenge successMessage.md
				created, err = ensureMarkdownFile(challengePath, "successMessage.md")
				if err != nil {
					result.Errors = append(result.Errors, fmt.Errorf("%s: %v", challengeRelPath, err))
				} else if created {
					result.FilesCreated++
				}
			}
		}
	}

	if len(result.Errors) > 0 {
		return result, fmt.Errorf("encountered %d error(s) during markdown generation", len(result.Errors))
	}

	return result, nil
}

// ensureMarkdownFile ensures a markdown file exists
// If the file doesn't exist, creates a placeholder file
// Returns (created bool, error) - created=true if file was created
func ensureMarkdownFile(dir, filename string) (bool, error) {
	path := filepath.Join(dir, filename)

	// Check if markdown file already exists
	_, err := os.Stat(path)
	if err == nil {
		// File exists, nothing to do
		return false, nil
	}

	// File doesn't exist, create placeholder
	content := "Placeholder\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return false, fmt.Errorf("failed to create %s: %w", filename, err)
	}

	return true, nil
}
