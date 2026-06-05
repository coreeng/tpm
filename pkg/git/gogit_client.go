package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"gopkg.in/yaml.v3"
)

// GoGitClient implements GitClient using go-git library
type GoGitClient struct {
	repo *git.Repository
}

// NewGoGitClient creates a new GoGitClient by opening the repository in the current directory
func NewGoGitClient() (*GoGitClient, error) {
	// Find the git repository starting from current directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	repo, err := git.PlainOpenWithOptions(cwd, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	return &GoGitClient{repo: repo}, nil
}

// ValidateRef checks if a git ref exists
func (c *GoGitClient) ValidateRef(ref string) error {
	_, err := c.repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return fmt.Errorf("invalid git ref '%s': %w", ref, err)
	}
	return nil
}

// IsGitRepo checks if the current directory is inside a git repository
func (c *GoGitClient) IsGitRepo() bool {
	return c.repo != nil
}

// GetFileContent reads a file from a specific git ref without checking it out
func (c *GoGitClient) GetFileContent(ref, filePath string) ([]byte, error) {
	// Resolve the ref to a hash
	hash, err := c.repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ref %s: %w", ref, err)
	}

	// Get the commit
	commit, err := c.repo.CommitObject(*hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}

	// Get the tree
	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}

	// Get the file
	file, err := tree.File(filePath)
	if err != nil {
		return nil, fmt.Errorf("file not found in ref %s: %s", ref, filePath)
	}

	// Read file contents
	contents, err := file.Contents()
	if err != nil {
		return nil, fmt.Errorf("failed to read file contents: %w", err)
	}

	return []byte(contents), nil
}

// ListFiles lists files matching a pattern in a git ref
func (c *GoGitClient) ListFiles(ref, pattern string) ([]string, error) {
	// Resolve the ref to a hash
	hash, err := c.repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ref %s: %w", ref, err)
	}

	// Get the commit
	commit, err := c.repo.CommitObject(*hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}

	// Get the tree
	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}

	// Collect matching files
	files := make([]string, 0)
	err = tree.Files().ForEach(func(f *object.File) error {
		if matchPattern(f.Name, pattern) {
			files = append(files, f.Name)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate files: %w", err)
	}

	return files, nil
}

// matchPattern performs simple wildcard pattern matching
// Supports * as wildcard for path components
func matchPattern(path, pattern string) bool {
	pathParts := strings.Split(path, "/")
	patternParts := strings.Split(pattern, "/")

	if len(pathParts) != len(patternParts) {
		return false
	}

	for i := range pathParts {
		if patternParts[i] == "*" {
			continue
		}
		// Handle patterns like "chapter.y*l" (wildcard within filename)
		if strings.Contains(patternParts[i], "*") {
			matched, _ := filepath.Match(patternParts[i], pathParts[i])
			if !matched {
				return false
			}
		} else if pathParts[i] != patternParts[i] {
			return false
		}
	}
	return true
}

// FindModulesAtRef finds all modules at a specific git ref
func (c *GoGitClient) FindModulesAtRef(ref string) ([]string, error) {
	// Resolve the ref to a hash
	hash, err := c.repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ref %s: %w", ref, err)
	}

	// Get the commit
	commit, err := c.repo.CommitObject(*hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}

	// Get the tree
	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}

	// Find all module files
	moduleMap := make(map[string]bool)
	err = tree.Files().ForEach(func(f *object.File) error {
		// Look for files that match pattern: <module-name>/module/module.{yaml,yml}
		if strings.Contains(f.Name, "/module/module.yaml") || strings.Contains(f.Name, "/module/module.yml") {
			parts := strings.Split(f.Name, "/")
			if len(parts) >= 3 && parts[1] == "module" {
				moduleName := parts[0]
				moduleMap[moduleName] = true
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate files: %w", err)
	}

	modules := make([]string, 0, len(moduleMap))
	for mod := range moduleMap {
		modules = append(modules, mod)
	}
	return modules, nil
}

// CollectCodesAtRef collects all codes from all modules at a specific git ref
func (c *GoGitClient) CollectCodesAtRef(ref string) (map[string]CodeInfo, error) {
	codes := make(map[string]CodeInfo)

	modules, err := c.FindModulesAtRef(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to find modules: %w", err)
	}

	for _, moduleName := range modules {
		moduleCodes, err := c.collectModuleCodesAtRef(ref, moduleName)
		if err != nil {
			// Log warning but continue with other modules
			fmt.Printf("Warning: failed to collect codes from %s at %s: %v\n", moduleName, ref, err)
			continue
		}

		for code, info := range moduleCodes {
			codes[code] = info
		}
	}

	return codes, nil
}

// collectModuleCodesAtRef collects codes from a single module at a specific ref
func (c *GoGitClient) collectModuleCodesAtRef(ref, moduleName string) (map[string]CodeInfo, error) {
	// Use the new hierarchical collection with parent tracking
	return c.collectModuleCodesAtRefNew(ref, moduleName)
}

// collectModuleCodesAtRefNew is the hierarchical version with parent tracking
func (c *GoGitClient) collectModuleCodesAtRefNew(ref, moduleName string) (map[string]CodeInfo, error) {
	codes := make(map[string]CodeInfo)

	// Try both .yaml and .yml extensions
	moduleFile := filepath.Join(moduleName, "module", "module.yaml")
	content, err := c.GetFileContent(ref, moduleFile)
	if err != nil {
		moduleFile = filepath.Join(moduleName, "module", "module.yml")
		content, err = c.GetFileContent(ref, moduleFile)
		if err != nil {
			return nil, fmt.Errorf("module file not found: %w", err)
		}
	}

	// Parse module file to get code
	var moduleData map[string]interface{}
	if err := yaml.Unmarshal(content, &moduleData); err != nil {
		return nil, fmt.Errorf("failed to parse module file: %w", err)
	}

	var moduleCode string
	if code, ok := moduleData["code"].(string); ok && code != "" {
		moduleCode = code
		codes[code] = CodeInfo{
			Code:       code,
			EntityType: "module",
			FilePath:   moduleFile,
			ModuleName: moduleName,
			ParentCode: "", // Module has no parent
			ParentType: "",
		}
	}

	// Collect codes from chapters (parent: module)
	chapterPattern := filepath.Join(moduleName, "module", "*", "chapter.y*l")
	chapterFiles, err := c.ListFiles(ref, chapterPattern)
	if err == nil {
		for _, chapterFile := range chapterFiles {
			chapterCodes, err := c.collectChapterCodesNew(ref, chapterFile, moduleName, moduleCode)
			if err != nil {
				continue
			}
			for code, info := range chapterCodes {
				codes[code] = info
			}
		}
	}

	return codes, nil
}

// collectChapterCodesNew collects chapter and its children with parent tracking
func (c *GoGitClient) collectChapterCodesNew(ref, chapterFile, moduleName, moduleCode string) (map[string]CodeInfo, error) {
	codes := make(map[string]CodeInfo)

	content, err := c.GetFileContent(ref, chapterFile)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := yaml.Unmarshal(content, &data); err != nil {
		return nil, err
	}

	// Get chapter directory path
	chapterDir := filepath.Dir(chapterFile)

	// Collect chapter code
	var chapterCode string
	if code, ok := data["code"].(string); ok && code != "" {
		chapterCode = code
		codes[code] = CodeInfo{
			Code:       code,
			EntityType: "chapter",
			FilePath:   chapterFile,
			ModuleName: moduleName,
			ParentCode: moduleCode,
			ParentType: "module",
		}
	}

	// Collect MCQ codes (parent: chapter)
	if mcqsRaw, ok := data["multipleChoiceAssessments"]; ok {
		if mcqs, ok := mcqsRaw.([]interface{}); ok {
			for _, mcqRaw := range mcqs {
				if mcq, ok := mcqRaw.(map[string]interface{}); ok {
					var mcqCode string
					if code, ok := mcq["code"].(string); ok && code != "" {
						mcqCode = code
						codes[code] = CodeInfo{
							Code:       code,
							EntityType: "multiple-choice-assessment",
							FilePath:   chapterFile,
							ModuleName: moduleName,
							ParentCode: chapterCode,
							ParentType: "chapter",
						}
					}

					// Collect question codes (parent: MCA)
					if questionsRaw, ok := mcq["questions"]; ok {
						if questions, ok := questionsRaw.([]interface{}); ok {
							for _, questionRaw := range questions {
								if question, ok := questionRaw.(map[string]interface{}); ok {
									if code, ok := question["code"].(string); ok && code != "" {
										codes[code] = CodeInfo{
											Code:       code,
											EntityType: "question",
											FilePath:   chapterFile,
											ModuleName: moduleName,
											ParentCode: mcqCode,
											ParentType: "multiple-choice-assessment",
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Collect sections (parent: chapter)
	sectionPattern := filepath.Join(chapterDir, "sections", "*", "section.y*l")
	sectionFiles, err := c.ListFiles(ref, sectionPattern)
	if err == nil {
		for _, sectionFile := range sectionFiles {
			if code, err := c.collectSingleCodeNew(ref, sectionFile, moduleName, "section", chapterCode, "chapter"); err == nil {
				codes[code.Code] = code
			}
		}
	}

	// Collect assessments (parent: chapter)
	assessmentPattern := filepath.Join(chapterDir, "assessments", "*", "assessment.y*l")
	assessmentFiles, err := c.ListFiles(ref, assessmentPattern)
	if err == nil {
		for _, assessmentFile := range assessmentFiles {
			assessmentCodes, err := c.collectAssessmentCodesNew(ref, assessmentFile, moduleName, chapterCode)
			if err != nil {
				continue
			}
			for code, info := range assessmentCodes {
				codes[code] = info
			}
		}
	}

	return codes, nil
}

// collectAssessmentCodesNew collects assessment and challenges with parent tracking
func (c *GoGitClient) collectAssessmentCodesNew(ref, assessmentFile, moduleName, chapterCode string) (map[string]CodeInfo, error) {
	codes := make(map[string]CodeInfo)

	// Get assessment code
	var assessmentCode string
	if code, err := c.collectSingleCodeNew(ref, assessmentFile, moduleName, "assessment", chapterCode, "chapter"); err == nil {
		assessmentCode = code.Code
		codes[code.Code] = code
	}

	// Get assessment directory
	assessmentDir := filepath.Dir(assessmentFile)

	// Collect challenges (parent: assessment)
	challengePattern := filepath.Join(assessmentDir, "*", "challenge.y*l")
	challengeFiles, err := c.ListFiles(ref, challengePattern)
	if err == nil {
		for _, challengeFile := range challengeFiles {
			challengeCodes, err := c.collectChallengeCodesNew(ref, challengeFile, moduleName, assessmentCode)
			if err != nil {
				continue
			}
			for code, info := range challengeCodes {
				codes[code] = info
			}
		}
	}

	return codes, nil
}

// collectChallengeCodesNew collects challenge and goals with parent tracking
func (c *GoGitClient) collectChallengeCodesNew(ref, challengeFile, moduleName, assessmentCode string) (map[string]CodeInfo, error) {
	codes := make(map[string]CodeInfo)

	content, err := c.GetFileContent(ref, challengeFile)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := yaml.Unmarshal(content, &data); err != nil {
		return nil, err
	}

	// Collect challenge code
	var challengeCode string
	if code, ok := data["code"].(string); ok && code != "" {
		challengeCode = code
		codes[code] = CodeInfo{
			Code:       code,
			EntityType: "challenge",
			FilePath:   challengeFile,
			ModuleName: moduleName,
			ParentCode: assessmentCode,
			ParentType: "assessment",
		}
	}

	// Collect goal codes (parent: challenge)
	if goalsRaw, ok := data["goals"]; ok {
		if goals, ok := goalsRaw.([]interface{}); ok {
			for _, goalRaw := range goals {
				if goal, ok := goalRaw.(map[string]interface{}); ok {
					if code, ok := goal["code"].(string); ok && code != "" {
						codes[code] = CodeInfo{
							Code:       code,
							EntityType: "goal",
							FilePath:   challengeFile,
							ModuleName: moduleName,
							ParentCode: challengeCode,
							ParentType: "challenge",
						}
					}
				}
			}
		}
	}

	return codes, nil
}

// collectSingleCodeNew collects a single code from a file with parent info
func (c *GoGitClient) collectSingleCodeNew(ref, filePath, moduleName, entityType, parentCode, parentType string) (CodeInfo, error) {
	content, err := c.GetFileContent(ref, filePath)
	if err != nil {
		return CodeInfo{}, err
	}

	var data map[string]interface{}
	if err := yaml.Unmarshal(content, &data); err != nil {
		return CodeInfo{}, err
	}

	if code, ok := data["code"].(string); ok && code != "" {
		return CodeInfo{
			Code:       code,
			EntityType: entityType,
			FilePath:   filePath,
			ModuleName: moduleName,
			ParentCode: parentCode,
			ParentType: parentType,
		}, nil
	}

	return CodeInfo{}, fmt.Errorf("no code found in %s", filePath)
}
