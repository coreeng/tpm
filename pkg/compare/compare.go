package compare

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/coreeng/tpm/pkg/artifact"
	"github.com/coreeng/tpm/pkg/module"
	"gopkg.in/yaml.v3"
)

const maxExtractedArchiveFileBytes = 100 << 20

type BreakingPolicy string

const (
	BreakingPolicyError  BreakingPolicy = "error"
	BreakingPolicyWarn   BreakingPolicy = "warn"
	BreakingPolicyIgnore BreakingPolicy = "ignore"
)

type Location struct {
	Input string
	Path  string
	Ref   string
}

type CodeInfo struct {
	Code       string
	EntityType string
	FilePath   string
	ParentCode string
	ParentType string
}

type Change struct {
	Code       string
	EntityType string
	Old        CodeInfo
	New        CodeInfo
}

type Report struct {
	OldLocation string
	NewLocation string
	OldCount    int
	NewCount    int
	Added       []CodeInfo
	Removed     []CodeInfo
	Moved       []Change
}

func (r *Report) HasBreakingChanges() bool {
	return len(r.Removed) > 0 || len(r.Moved) > 0
}

func ValidateBreakingPolicy(policy string) (BreakingPolicy, error) {
	switch BreakingPolicy(policy) {
	case BreakingPolicyError, BreakingPolicyWarn, BreakingPolicyIgnore:
		return BreakingPolicy(policy), nil
	default:
		return "", fmt.Errorf("breaking-policy must be one of: error, warn, ignore")
	}
}

func Compare(oldArg, newArg string) (*Report, error) {
	oldLocation, err := ParseLocation(oldArg)
	if err != nil {
		return nil, err
	}
	newLocation, err := ParseLocation(newArg)
	if err != nil {
		return nil, err
	}

	oldCodes, err := Collect(oldLocation)
	if err != nil {
		return nil, fmt.Errorf("collect old location %s: %w", oldArg, err)
	}
	newCodes, err := Collect(newLocation)
	if err != nil {
		return nil, fmt.Errorf("collect new location %s: %w", newArg, err)
	}

	report := &Report{
		OldLocation: oldArg,
		NewLocation: newArg,
		OldCount:    len(oldCodes),
		NewCount:    len(newCodes),
	}

	for code, oldInfo := range oldCodes {
		newInfo, ok := newCodes[code]
		if !ok {
			report.Removed = append(report.Removed, oldInfo)
			continue
		}
		if oldInfo.ParentCode != newInfo.ParentCode || oldInfo.ParentType != newInfo.ParentType {
			report.Moved = append(report.Moved, Change{Code: code, EntityType: oldInfo.EntityType, Old: oldInfo, New: newInfo})
		}
	}
	for code, newInfo := range newCodes {
		if _, ok := oldCodes[code]; !ok {
			report.Added = append(report.Added, newInfo)
		}
	}

	sortCodeInfos(report.Added)
	sortCodeInfos(report.Removed)
	sort.Slice(report.Moved, func(i, j int) bool {
		return report.Moved[i].Code < report.Moved[j].Code
	})
	return report, nil
}

func ParseLocation(input string) (Location, error) {
	if strings.TrimSpace(input) == "" {
		return Location{}, fmt.Errorf("location is required")
	}
	location := Location{Input: input, Path: input}
	if index := strings.LastIndex(input, "@"); index > 0 && index < len(input)-1 {
		location.Path = input[:index]
		location.Ref = input[index+1:]
	}
	if strings.TrimSpace(location.Path) == "" {
		return Location{}, fmt.Errorf("location path is required")
	}
	return location, nil
}

func Collect(location Location) (map[string]CodeInfo, error) {
	if location.Ref == "" {
		return collectLocal(location.Path)
	}
	localPath, cleanup, err := materializeGitLocation(location)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	return collectLocal(localPath)
}

func collectLocal(path string) (map[string]CodeInfo, error) {
	clean := filepath.Clean(path)
	info, err := os.Stat(clean)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		if looksLikeBuiltArtifact(clean) {
			return collectBuiltFile(clean)
		}
		return nil, fmt.Errorf("%s is a file; compare accepts built module.yaml files or module directories", path)
	}

	directModuleFile := findDirectModuleFile(clean)
	if directModuleFile != "" && looksLikeBuiltArtifact(directModuleFile) {
		return collectBuiltFile(directModuleFile)
	}
	if _, err := module.ResolvePath(clean); err == nil {
		return collectSourcePath(clean)
	}
	artifactPath, err := artifact.ResolveModuleArtifactPath(clean)
	if err == nil {
		return collectBuiltFile(artifactPath)
	}
	return nil, fmt.Errorf("%s is not a module source directory or built module artifact", path)
}

func collectSourcePath(path string) (map[string]CodeInfo, error) {
	mod, resolved, err := module.LoadPath(path)
	if err != nil {
		return nil, err
	}
	codes := map[string]CodeInfo{}
	collectSourceModule(codes, mod, resolved)
	return codes, nil
}

func collectSourceModule(codes map[string]CodeInfo, mod *module.Module, resolved module.ResolvedPath) {
	if mod.Code != "" {
		codes[mod.Code] = CodeInfo{Code: mod.Code, EntityType: "module", FilePath: resolved.ModuleFilePath}
	}
	for _, chapter := range mod.Chapters {
		if chapter.Code != "" {
			codes[chapter.Code] = CodeInfo{Code: chapter.Code, EntityType: "chapter", FilePath: chapter.FilePath, ParentCode: mod.Code, ParentType: "module"}
		}
		for _, section := range chapter.Sections {
			if section.Code != "" {
				codes[section.Code] = CodeInfo{Code: section.Code, EntityType: "section", FilePath: section.FilePath, ParentCode: chapter.Code, ParentType: "chapter"}
			}
		}
		for _, assessment := range chapter.Assessments {
			if assessment.Code != "" {
				codes[assessment.Code] = CodeInfo{Code: assessment.Code, EntityType: "lab", FilePath: assessment.FilePath, ParentCode: chapter.Code, ParentType: "chapter"}
			}
			for _, challenge := range assessment.Challenges {
				if challenge.Code != "" {
					codes[challenge.Code] = CodeInfo{Code: challenge.Code, EntityType: "challenge", FilePath: challenge.FilePath, ParentCode: assessment.Code, ParentType: "lab"}
				}
				for _, goal := range challenge.Goals {
					if goal.Code != "" {
						codes[goal.Code] = CodeInfo{Code: goal.Code, EntityType: "goal", FilePath: challenge.FilePath, ParentCode: challenge.Code, ParentType: "challenge"}
					}
				}
			}
		}
		for _, quiz := range chapter.MultipleChoiceAssessments {
			if quiz.Code != "" {
				codes[quiz.Code] = CodeInfo{Code: quiz.Code, EntityType: "quiz", FilePath: chapter.FilePath, ParentCode: chapter.Code, ParentType: "chapter"}
			}
			for _, question := range quiz.Questions {
				if question.Code != "" {
					codes[question.Code] = CodeInfo{Code: question.Code, EntityType: "question", FilePath: chapter.FilePath, ParentCode: quiz.Code, ParentType: "quiz"}
				}
			}
		}
	}
}

func collectBuiltFile(path string) (map[string]CodeInfo, error) {
	// #nosec G304 -- compare intentionally reads a module artifact path supplied by the local CLI user.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var mod module.BuiltModule
	if err := yaml.Unmarshal(data, &mod); err != nil {
		return nil, err
	}
	codes := map[string]CodeInfo{}
	if mod.Code != "" {
		codes[mod.Code] = CodeInfo{Code: mod.Code, EntityType: "module", FilePath: path}
	}
	for _, chapter := range mod.Chapters {
		if chapter.Code != "" {
			codes[chapter.Code] = CodeInfo{Code: chapter.Code, EntityType: "chapter", FilePath: path, ParentCode: mod.Code, ParentType: "module"}
		}
		for _, section := range chapter.Sections {
			if section.Code != "" {
				codes[section.Code] = CodeInfo{Code: section.Code, EntityType: "section", FilePath: path, ParentCode: chapter.Code, ParentType: "chapter"}
			}
		}
		for _, assessment := range chapter.Assessments {
			if assessment.Code != "" {
				codes[assessment.Code] = CodeInfo{Code: assessment.Code, EntityType: "lab", FilePath: path, ParentCode: chapter.Code, ParentType: "chapter"}
			}
			for _, challenge := range assessment.Challenges {
				if challenge.Code != "" {
					codes[challenge.Code] = CodeInfo{Code: challenge.Code, EntityType: "challenge", FilePath: path, ParentCode: assessment.Code, ParentType: "lab"}
				}
				for _, goal := range challenge.Goals {
					if goal.Code != "" {
						codes[goal.Code] = CodeInfo{Code: goal.Code, EntityType: "goal", FilePath: path, ParentCode: challenge.Code, ParentType: "challenge"}
					}
				}
			}
		}
		for _, quiz := range chapter.MultipleChoiceAssessments {
			if quiz.Code != "" {
				codes[quiz.Code] = CodeInfo{Code: quiz.Code, EntityType: "quiz", FilePath: path, ParentCode: chapter.Code, ParentType: "chapter"}
			}
			for _, question := range quiz.Questions {
				if question.Code != "" {
					codes[question.Code] = CodeInfo{Code: question.Code, EntityType: "question", FilePath: path, ParentCode: quiz.Code, ParentType: "quiz"}
				}
			}
		}
	}
	return codes, nil
}

func looksLikeBuiltArtifact(path string) bool {
	// #nosec G304 -- compare intentionally probes a module artifact path supplied by the local CLI user.
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var root map[string]any
	if err := yaml.Unmarshal(data, &root); err != nil {
		return false
	}
	_, ok := root["chapters"]
	return ok
}

func findDirectModuleFile(dir string) string {
	for _, name := range []string{"module.yaml", "module.yml"} {
		path := filepath.Join(dir, name)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}
	return ""
}

func materializeGitLocation(location Location) (string, func(), error) {
	repoRoot, err := gitOutput("rev-parse", "--show-toplevel")
	if err != nil {
		return "", nil, fmt.Errorf("not in a git repository: %w", err)
	}
	repoRoot = strings.TrimSpace(repoRoot)
	relPath, err := gitRelativePath(repoRoot, location.Path)
	if err != nil {
		return "", nil, err
	}
	tempDir, err := os.MkdirTemp("", "tpm-compare-")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() {
		_ = os.RemoveAll(tempDir)
	}
	// #nosec G204 -- git is a fixed executable; ref/path are passed as argv and not through a shell.
	cmd := exec.Command("git", "-C", repoRoot, "archive", "--format=tar", location.Ref, "--", relPath)
	tarball, err := cmd.Output()
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("archive %s at %s: %w", relPath, location.Ref, err)
	}
	if err := untar(bytes.NewReader(tarball), tempDir); err != nil {
		cleanup()
		return "", nil, err
	}
	if relPath == "." {
		return tempDir, cleanup, nil
	}
	return filepath.Join(tempDir, filepath.FromSlash(relPath)), cleanup, nil
}

func gitRelativePath(repoRoot, path string) (string, error) {
	if path == "." {
		return ".", nil
	}
	abs := path
	if !filepath.IsAbs(abs) {
		var err error
		abs, err = filepath.Abs(path)
		if err != nil {
			return "", err
		}
	}
	rel, err := filepath.Rel(repoRoot, abs)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("%s is outside git repository %s", path, repoRoot)
	}
	return filepath.ToSlash(filepath.Clean(rel)), nil
}

func gitOutput(args ...string) (string, error) {
	// #nosec G204 -- git is a fixed executable; args are passed as argv and not through a shell.
	output, err := exec.Command("git", args...).Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func untar(reader io.Reader, dir string) error {
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		target, skip, err := safeTarTarget(dir, header.Name)
		if err != nil {
			return err
		}
		if skip {
			continue
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0700); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
				return err
			}
			// #nosec G304 -- target is validated by safeTarTarget to remain under dir.
			file, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
			if err != nil {
				return err
			}
			limitedReader := &io.LimitedReader{R: tarReader, N: maxExtractedArchiveFileBytes + 1}
			// #nosec G110 -- extraction is bounded by maxExtractedArchiveFileBytes.
			written, err := io.Copy(file, limitedReader)
			if err != nil {
				_ = file.Close()
				return err
			}
			if written > maxExtractedArchiveFileBytes {
				_ = file.Close()
				return fmt.Errorf("archive entry %s exceeds %d bytes", header.Name, maxExtractedArchiveFileBytes)
			}
			if err := file.Close(); err != nil {
				return err
			}
		}
	}
}

func safeTarTarget(rootDir, name string) (string, bool, error) {
	if name == "" || name == "." {
		return "", true, nil
	}
	if strings.ContainsRune(name, 0) {
		return "", false, fmt.Errorf("archive entry %q contains a null byte", name)
	}
	if strings.Contains(name, "\\") {
		return "", false, fmt.Errorf("archive entry %q contains a backslash path separator", name)
	}
	if path.IsAbs(name) || filepath.IsAbs(name) {
		return "", false, fmt.Errorf("archive entry %q is absolute", name)
	}

	cleanName := path.Clean(name)
	if cleanName == "." {
		return "", true, nil
	}
	if cleanName == ".." || strings.HasPrefix(cleanName, "../") {
		return "", false, fmt.Errorf("archive entry %q escapes extraction root", name)
	}

	rootAbs, err := filepath.Abs(rootDir)
	if err != nil {
		return "", false, err
	}
	targetAbs, err := filepath.Abs(filepath.Join(rootAbs, filepath.FromSlash(cleanName)))
	if err != nil {
		return "", false, err
	}
	rel, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil {
		return "", false, err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", false, fmt.Errorf("archive entry %q escapes extraction root", name)
	}
	return targetAbs, false, nil
}

func sortCodeInfos(infos []CodeInfo) {
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Code < infos[j].Code
	})
}
