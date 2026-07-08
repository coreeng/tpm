package moduleinit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ModuleScaffoldOptions struct {
	Name string
}

func ScaffoldModuleSkeleton(dir string, opts ModuleScaffoldOptions) error {
	if err := validateName(opts.Name); err != nil {
		return err
	}

	target := filepath.Join(dir, opts.Name)
	if err := ensureEmptyTarget(target); err != nil {
		return err
	}

	files := map[string]string{
		"module/module.yaml":                       moduleYAML(opts.Name),
		"module/description.md":                    moduleDescription(opts.Name),
		"module/01-getting-started/chapter.yml":    chapterYAML,
		"module/01-getting-started/description.md": chapterDescription,
	}
	for name, contents := range files {
		path := filepath.Join(target, name)
		if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
			return fmt.Errorf("create parent for %s: %w", path, err)
		}
		if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}
	return nil
}

func validateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("module name is required")
	}
	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		return fmt.Errorf("module name %q must not start or end with '-'", name)
	}
	for _, r := range name {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' {
			continue
		}
		return fmt.Errorf("module name %q must contain only lowercase letters, numbers, and '-'", name)
	}
	return nil
}

func ensureEmptyTarget(dir string) error {
	entries, err := os.ReadDir(dir)
	if err == nil {
		if len(entries) > 0 {
			return fmt.Errorf("target directory %s is not empty", dir)
		}
		return nil
	}
	if os.IsNotExist(err) {
		return nil
	}
	return fmt.Errorf("read target directory %s: %w", dir, err)
}

func moduleYAML(name string) string {
	return fmt.Sprintf(`code: %s
title: %s
shortDescription: Local lab module
bannerImage: https://storage.googleapis.com/cdn-training-platform/local-lab-banner.png
level: BEGINNER
tags:
  - local-lab
`, name, titleFromName(name))
}

func moduleDescription(name string) string {
	return fmt.Sprintf(`# %s

Local lab module.
`, titleFromName(name))
}

func titleFromName(name string) string {
	parts := strings.Split(name, "-")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

const chapterYAML = `code: getting-started
title: Getting Started
shortDescription: Start building local labs.
isDraft: false
`

const chapterDescription = `# Getting Started

Start building local labs for this module.
`
