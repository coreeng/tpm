package lab

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

func Load(path string) (*Lab, error) {
	runtimePath := filepath.Clean(path)
	if err := validateRuntimeDirs(runtimePath); err != nil {
		return nil, err
	}

	metadataPath := filepath.Join(runtimePath, "lab.yaml")
	if dirExists(runtimePath) && fileExists(metadataPath) {
		return loadStandalone(runtimePath, metadataPath)
	}

	metadataRoot, err := moduleMetadataRoot(runtimePath)
	if err != nil {
		return nil, err
	}
	return loadModuleBacked(metadataRoot, runtimePath)
}

func loadStandalone(rootPath, metadataPath string) (*Lab, error) {
	var loaded Lab
	if err := loadYAML(metadataPath, &loaded); err != nil {
		return nil, err
	}
	loaded.Format = FormatStandalone
	loaded.RootPath = rootPath
	loaded.RuntimePath = rootPath
	loaded.MetadataPath = metadataPath
	setRuntimePaths(&loaded, rootPath)
	return &loaded, nil
}

func loadModuleBacked(metadataRoot, runtimePath string) (*Lab, error) {
	metadataPath := filepath.Join(metadataRoot, "assessment.yaml")
	var loaded Lab
	if err := loadYAML(metadataPath, &loaded); err != nil {
		return nil, err
	}
	if description, err := readOptionalText(filepath.Join(metadataRoot, "description.md")); err == nil {
		loaded.Description = description
	}
	challenges, err := loadChallenges(metadataRoot)
	if err != nil {
		return nil, err
	}
	loaded.Format = FormatModuleBacked
	loaded.RootPath = metadataRoot
	loaded.RuntimePath = runtimePath
	loaded.MetadataPath = metadataPath
	loaded.Challenges = challenges
	setRuntimePaths(&loaded, runtimePath)
	return &loaded, nil
}

func validateRuntimeDirs(runtimePath string) error {
	for _, name := range []string{"starter-content", "solution", "validator"} {
		path := filepath.Join(runtimePath, name)
		if !dirExists(path) {
			return fmt.Errorf("required runtime directory %q not found: %w", path, os.ErrNotExist)
		}
	}
	return nil
}

func setRuntimePaths(loaded *Lab, runtimePath string) {
	loaded.StarterPath = filepath.Join(runtimePath, "starter-content")
	loaded.SolutionPath = filepath.Join(runtimePath, "solution")
	loaded.ValidatorPath = filepath.Join(runtimePath, "validator")
}

func moduleMetadataRoot(runtimePath string) (string, error) {
	labName := filepath.Base(runtimePath)
	chapterPath := filepath.Dir(runtimePath)
	chapterName := filepath.Base(chapterPath)
	assessmentsPath := filepath.Dir(chapterPath)
	if filepath.Base(assessmentsPath) != "assessments" {
		return "", fmt.Errorf("runtime path %q is not under an assessments directory", runtimePath)
	}
	moduleRoot := filepath.Dir(assessmentsPath)
	return filepath.Join(moduleRoot, "module", chapterName, "assessments", labName), nil
}

func loadChallenges(metadataRoot string) ([]Challenge, error) {
	entries, err := os.ReadDir(metadataRoot)
	if err != nil {
		return nil, err
	}

	type loadedChallenge struct {
		dirName   string
		challenge Challenge
	}

	var loaded []loadedChallenge
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		challengePath := findChallengeFile(filepath.Join(metadataRoot, entry.Name()))
		if !fileExists(challengePath) {
			continue
		}
		var challenge Challenge
		if err := loadYAML(challengePath, &challenge); err != nil {
			return nil, err
		}
		challengeRoot := filepath.Dir(challengePath)
		if description, err := readOptionalText(filepath.Join(challengeRoot, "description.md")); err == nil {
			challenge.Description = description
		}
		if successMessage, err := readOptionalText(filepath.Join(challengeRoot, "successMessage.md")); err == nil {
			challenge.SuccessMessage = successMessage
		}
		loaded = append(loaded, loadedChallenge{dirName: entry.Name(), challenge: challenge})
	}
	sort.Slice(loaded, func(i, j int) bool {
		return loaded[i].dirName < loaded[j].dirName
	})

	var challenges []Challenge
	for _, item := range loaded {
		challenges = append(challenges, item.challenge)
	}
	return challenges, nil
}

func findChallengeFile(path string) string {
	for _, name := range []string{"challenge.yaml", "challenge.yml"} {
		candidate := filepath.Join(path, name)
		if fileExists(candidate) {
			return candidate
		}
	}
	return ""
}

func loadYAML(path string, into any) error {
	// #nosec G304 -- lab loading intentionally reads local lab YAML selected by the CLI user.
	contents, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(contents, into)
}

func readOptionalText(path string) (string, error) {
	// #nosec G304 -- lab loading intentionally reads optional markdown from the selected local lab tree.
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
