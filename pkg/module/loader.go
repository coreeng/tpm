package module

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/coreeng/tpm/pkg/pathutil"
	"gopkg.in/yaml.v3"
)

// LoadModule loads a module and all its components from disk
func LoadModule(rootDir, moduleName string) (*Module, error) {
	modulePath := GetModulePath(rootDir, moduleName)
	moduleFilePath := GetModuleFilePath(rootDir, moduleName)

	// Check if module file exists
	if !pathutil.FileExists(moduleFilePath) {
		return nil, fmt.Errorf("module file not found at %s", moduleFilePath)
	}

	// Load module metadata
	var module Module
	if err := loadYAML(moduleFilePath, &module); err != nil {
		return nil, fmt.Errorf("failed to load module file: %w", err)
	}
	module.FilePath = moduleFilePath

	// Load all chapters
	chapters, err := loadChapters(modulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chapters: %w", err)
	}
	module.Chapters = chapters

	return &module, nil
}

// loadChapters finds and loads all chapters in the module directory
func loadChapters(modulePath string) ([]Chapter, error) {
	var chapters []Chapter

	entries, err := os.ReadDir(modulePath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip non-chapter directories
		name := entry.Name()
		if name == "p2p" || name[0] == '.' {
			continue
		}

		chapterPath := filepath.Join(modulePath, name)

		// Try to find chapter.yml or chapter.yaml
		chapterFile := findFile(chapterPath, "chapter")
		if chapterFile == "" {
			continue // Not a chapter directory
		}

		// Load chapter
		var chapter Chapter
		if err := loadYAML(chapterFile, &chapter); err != nil {
			return nil, fmt.Errorf("failed to load chapter %s: %w", chapterFile, err)
		}
		chapter.FilePath = chapterFile

		// Load sections (directly from chapter directory, not from a sections/ subdirectory)
		sections, err := loadSections(chapterPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load sections for chapter %s: %w", name, err)
		}
		chapter.Sections = sections

		// Load assessments
		assessmentsPath := filepath.Join(chapterPath, "assessments")
		if pathutil.DirExists(assessmentsPath) {
			assessments, err := loadAssessments(assessmentsPath)
			if err != nil {
				return nil, fmt.Errorf("failed to load assessments for chapter %s: %w", name, err)
			}
			chapter.Assessments = assessments
		}

		chapters = append(chapters, chapter)
	}

	// Sort chapters by their directory name (which starts with index)
	sort.Slice(chapters, func(i, j int) bool {
		return filepath.Base(filepath.Dir(chapters[i].FilePath)) < filepath.Base(filepath.Dir(chapters[j].FilePath))
	})

	return chapters, nil
}

// loadSections finds and loads all sections in a chapter directory
func loadSections(chapterPath string) ([]Section, error) {
	var sections []Section

	entries, err := os.ReadDir(chapterPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip special directories (assessments, p2p, hidden dirs)
		name := entry.Name()
		if name == "assessments" || name == "p2p" || name[0] == '.' {
			continue
		}

		sectionPath := filepath.Join(chapterPath, name)
		sectionFile := findFile(sectionPath, "section")
		if sectionFile == "" {
			continue
		}

		var section Section
		if err := loadYAML(sectionFile, &section); err != nil {
			return nil, fmt.Errorf("failed to load section %s: %w", sectionFile, err)
		}
		section.FilePath = sectionFile

		sections = append(sections, section)
	}

	// Sort sections by directory name
	sort.Slice(sections, func(i, j int) bool {
		return filepath.Base(filepath.Dir(sections[i].FilePath)) < filepath.Base(filepath.Dir(sections[j].FilePath))
	})

	return sections, nil
}

// loadAssessments finds and loads all assessments in an assessments directory
func loadAssessments(assessmentsPath string) ([]Assessment, error) {
	var assessments []Assessment

	entries, err := os.ReadDir(assessmentsPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		assessmentPath := filepath.Join(assessmentsPath, entry.Name())
		assessmentFile := findFile(assessmentPath, "assessment")
		if assessmentFile == "" {
			continue
		}

		var assessment Assessment
		if err := loadYAML(assessmentFile, &assessment); err != nil {
			return nil, fmt.Errorf("failed to load assessment %s: %w", assessmentFile, err)
		}
		assessment.FilePath = assessmentFile

		// Load challenges
		challenges, err := loadChallenges(assessmentPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load challenges for assessment %s: %w", entry.Name(), err)
		}
		assessment.Challenges = challenges

		assessments = append(assessments, assessment)
	}

	// Sort assessments by directory name
	sort.Slice(assessments, func(i, j int) bool {
		return filepath.Base(filepath.Dir(assessments[i].FilePath)) < filepath.Base(filepath.Dir(assessments[j].FilePath))
	})

	return assessments, nil
}

// loadChallenges finds and loads all challenges in an assessment directory
func loadChallenges(assessmentPath string) ([]Challenge, error) {
	var challenges []Challenge

	entries, err := os.ReadDir(assessmentPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		challengePath := filepath.Join(assessmentPath, entry.Name())
		challengeFile := findFile(challengePath, "challenge")
		if challengeFile == "" {
			continue
		}

		var challenge Challenge
		if err := loadYAML(challengeFile, &challenge); err != nil {
			return nil, fmt.Errorf("failed to load challenge %s: %w", challengeFile, err)
		}
		challenge.FilePath = challengeFile

		challenges = append(challenges, challenge)
	}

	// Sort challenges by directory name
	sort.Slice(challenges, func(i, j int) bool {
		return filepath.Base(filepath.Dir(challenges[i].FilePath)) < filepath.Base(filepath.Dir(challenges[j].FilePath))
	})

	return challenges, nil
}

// loadYAML reads a YAML file and unmarshals it into the provided struct
func loadYAML(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, v); err != nil {
		return err
	}

	return nil
}

// findFile looks for a file with the given base name and .yaml or .yml extension
func findFile(dir, baseName string) string {
	yamlPath := filepath.Join(dir, baseName+".yaml")
	if pathutil.FileExists(yamlPath) {
		return yamlPath
	}

	ymlPath := filepath.Join(dir, baseName+".yml")
	if pathutil.FileExists(ymlPath) {
		return ymlPath
	}

	return ""
}
