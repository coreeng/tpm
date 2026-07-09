package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/coreeng/tpm/pkg/module"
	"github.com/coreeng/tpm/pkg/pathutil"
)

// MergeDescriptions reads description.md files and loads them into entity descriptions
// Description fields should ONLY be specified in markdown files, not in YAML
// Returns an error if required markdown files are missing
func MergeDescriptions(mod *module.Module, modulePath string) ([]string, error) {
	var warnings []string
	var errors []string

	// Module level - description.md is REQUIRED
	desc, err := readDescriptionMd(modulePath)
	if err != nil {
		errors = append(errors, "module/module.yaml: missing required description.md file")
	} else if desc == "" {
		errors = append(errors, "module/module.yaml: description.md exists but is empty")
	} else {
		mod.Description = desc
	}

	// Chapters
	for i := range mod.Chapters {
		ch := &mod.Chapters[i]

		// Determine chapter directory name from file path
		chapterPath := filepath.Dir(ch.FilePath)
		relPath := pathutil.GetRelativeModulePath(ch.FilePath)

		// Chapter description.md is REQUIRED
		desc, err := readDescriptionMd(chapterPath)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: missing required description.md file", relPath))
		} else if desc == "" {
			errors = append(errors, fmt.Sprintf("%s: description.md exists but is empty", relPath))
		} else {
			ch.Description = desc
		}

		// Sections
		for j := range ch.Sections {
			sec := &ch.Sections[j]
			sectionPath := filepath.Dir(sec.FilePath)
			secRelPath := pathutil.GetRelativeModulePath(sec.FilePath)

			// Section description.md is REQUIRED
			desc, err := readDescriptionMd(sectionPath)
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: missing required description.md file", secRelPath))
			} else if desc == "" {
				errors = append(errors, fmt.Sprintf("%s: description.md exists but is empty", secRelPath))
			} else {
				sec.Description = desc
			}
		}

		// Interactive Assessments
		for j := range ch.Assessments {
			assessment := &ch.Assessments[j]
			assessmentPath := filepath.Dir(assessment.FilePath)
			assessRelPath := pathutil.GetRelativeModulePath(assessment.FilePath)

			// Assessment description.md is REQUIRED
			desc, err := readDescriptionMd(assessmentPath)
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: missing required description.md file", assessRelPath))
			} else if desc == "" {
				errors = append(errors, fmt.Sprintf("%s: description.md exists but is empty", assessRelPath))
			} else {
				assessment.Description = desc
			}

			// Challenges
			for k := range assessment.Challenges {
				challenge := &assessment.Challenges[k]
				challengePath := filepath.Dir(challenge.FilePath)
				challengeRelPath := pathutil.GetRelativeModulePath(challenge.FilePath)

				// Challenge description.md is REQUIRED
				desc, err := readDescriptionMd(challengePath)
				if err != nil {
					errors = append(errors, fmt.Sprintf("%s: missing required description.md file", challengeRelPath))
				} else if desc == "" {
					errors = append(errors, fmt.Sprintf("%s: description.md exists but is empty", challengeRelPath))
				} else {
					challenge.Description = desc
				}

				// Challenge successMessage.md is REQUIRED
				successMsg, err := readSuccessMessageMd(challengePath)
				if err != nil {
					errors = append(errors, fmt.Sprintf("%s: missing required successMessage.md file", challengeRelPath))
				} else if successMsg == "" {
					errors = append(errors, fmt.Sprintf("%s: successMessage.md exists but is empty", challengeRelPath))
				} else {
					challenge.SuccessMessage = successMsg
				}
			}
		}

		// Note: Multiple Choice Assessments don't have description.md files
		// as they're defined inline in chapter.yml
	}

	// Return error if any required files are missing
	if len(errors) > 0 {
		return warnings, fmt.Errorf("missing required markdown files:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return warnings, nil
}

// readDescriptionMd reads description.md from a directory
// Returns empty string if file doesn't exist (not an error)
func readDescriptionMd(dir string) (string, error) {
	path := filepath.Join(dir, "description.md")

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", err // Not found - caller should handle
	}

	// Read file content
	// #nosec G304 -- module builds intentionally read description.md from the selected local module tree.
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// readSuccessMessageMd reads successMessage.md from a directory
// Returns empty string if file doesn't exist (not an error)
func readSuccessMessageMd(dir string) (string, error) {
	path := filepath.Join(dir, "successMessage.md")

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", err // Not found - caller should handle
	}

	// Read file content
	// #nosec G304 -- module builds intentionally read successMessage.md from the selected local module tree.
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
