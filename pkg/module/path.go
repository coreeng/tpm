package module

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type ResolvedPath struct {
	Input          string
	RootPath       string
	SourcePath     string
	ModuleFilePath string
	Name           string
}

func ResolvePath(input string) (ResolvedPath, error) {
	if input == "" {
		return ResolvedPath{}, fmt.Errorf("module path is required")
	}
	clean := filepath.Clean(input)
	abs, err := filepath.Abs(clean)
	if err != nil {
		return ResolvedPath{}, fmt.Errorf("resolve %s: %w", input, err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return ResolvedPath{}, fmt.Errorf("inspect module path %s: %w", input, err)
	}
	if !info.IsDir() {
		return ResolvedPath{}, fmt.Errorf("module path %s must be a directory", input)
	}

	if moduleFile := findFile(filepath.Join(abs, "module"), "module"); moduleFile != "" {
		return resolved(input, abs, filepath.Join(abs, "module"), moduleFile), nil
	}
	if moduleFile := findFile(abs, "module"); moduleFile != "" {
		root := abs
		if filepath.Base(abs) == "module" {
			root = filepath.Dir(abs)
		}
		return resolved(input, root, abs, moduleFile), nil
	}

	return ResolvedPath{}, fmt.Errorf("module path %s must contain module/module.yaml, module/module.yml, module.yaml, or module.yml", input)
}

func resolved(input, rootPath, sourcePath, moduleFilePath string) ResolvedPath {
	return ResolvedPath{
		Input:          input,
		RootPath:       rootPath,
		SourcePath:     sourcePath,
		ModuleFilePath: moduleFilePath,
		Name:           filepath.Base(rootPath),
	}
}

func FindPaths(dir string) ([]ResolvedPath, error) {
	if dir == "" {
		return nil, fmt.Errorf("directory is required")
	}
	clean := filepath.Clean(dir)
	entries, err := os.ReadDir(clean)
	if err != nil {
		return nil, fmt.Errorf("read module list directory %s: %w", dir, err)
	}

	var modules []ResolvedPath
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name[0] == '.' || name == "node_modules" || name == "vendor" {
			continue
		}
		resolved, err := ResolvePath(filepath.Join(clean, name))
		if err == nil {
			modules = append(modules, resolved)
		}
	}
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Name < modules[j].Name
	})
	return modules, nil
}

func LoadPath(input string) (*Module, ResolvedPath, error) {
	resolved, err := ResolvePath(input)
	if err != nil {
		return nil, ResolvedPath{}, err
	}
	mod, err := loadModuleFromSourcePath(resolved.SourcePath, resolved.ModuleFilePath)
	if err != nil {
		return nil, ResolvedPath{}, err
	}
	return mod, resolved, nil
}
