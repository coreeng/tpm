package lab

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type RunState struct {
	LabPath            string    `yaml:"labPath"`
	RunID              string    `yaml:"runID"`
	SystemNamespace    string    `yaml:"systemNamespace"`
	WorkspaceNamespace string    `yaml:"workspaceNamespace"`
	ValidatorImageTag  string    `yaml:"validatorImageTag"`
	RegistryURL        string    `yaml:"registryURL"`
	RegistryUsername   string    `yaml:"registryUsername"`
	RegistryToken      string    `yaml:"registryToken"`
	ChartDir           string    `yaml:"chartDir,omitempty"`
	HelmReleaseName    string    `yaml:"helmReleaseName"`
	ChartURI           string    `yaml:"chartURI"`
	ChartVersion       string    `yaml:"chartVersion"`
	CreatedAt          time.Time `yaml:"createdAt"`
}

func StateDir(repoRoot string) string {
	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		return filepath.Join(home, ".config", "tpm", "labs")
	}
	return filepath.Join(repoRoot, ".build", "tpm", "labs")
}

func SaveState(stateDir string, state RunState) error {
	if err := validateRunID(state.RunID); err != nil {
		return err
	}
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return err
	}
	contents, err := yaml.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(stateDir, state.RunID+".yaml"), contents, 0644)
}

func LoadState(path string) (*RunState, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var state RunState
	if err := yaml.Unmarshal(contents, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func resolveState(opts Options) (*RunState, string, error) {
	if opts.RepoRoot == "" {
		opts.RepoRoot = "."
	}
	if opts.StateDir == "" {
		opts.StateDir = StateDir(opts.RepoRoot)
	}
	if opts.ID != "" {
		if err := validateRunID(opts.ID); err != nil {
			return nil, "", err
		}
		path := filepath.Join(opts.StateDir, opts.ID+".yaml")
		state, err := LoadState(path)
		if err != nil {
			return nil, "", fmt.Errorf("load lab run state: %w", err)
		}
		return state, path, nil
	}
	if opts.LabPath != "" {
		labPath, err := normalizeLabPath(opts.LabPath)
		if err != nil {
			return nil, "", err
		}
		state, err := FindLatestState(opts.StateDir, labPath)
		if err != nil {
			return nil, "", fmt.Errorf("find latest lab run state: %w", err)
		}
		if state == nil {
			return nil, "", fmt.Errorf("no lab run state found for %q", labPath)
		}
		if err := validateRunID(state.RunID); err != nil {
			return nil, "", err
		}
		return state, filepath.Join(opts.StateDir, state.RunID+".yaml"), nil
	}
	return nil, "", fmt.Errorf("lab run ID or lab path is required")
}

func validateRunID(id string) error {
	if id == "" {
		return fmt.Errorf("lab run ID is required")
	}
	if len(id) > 49 {
		return fmt.Errorf("lab run ID %q is too long; must be 49 characters or fewer", id)
	}
	if !isLowerAlnum(id[0]) || !isLowerAlnum(id[len(id)-1]) {
		return fmt.Errorf("lab run ID %q must start and end with a lowercase letter or digit", id)
	}
	for i := 0; i < len(id); i++ {
		if isLowerAlnum(id[i]) || id[i] == '-' {
			continue
		}
		return fmt.Errorf("lab run ID %q must contain only lowercase letters, digits, and hyphens", id)
	}
	return nil
}

func isLowerAlnum(char byte) bool {
	return char >= 'a' && char <= 'z' || char >= '0' && char <= '9'
}

func FindLatestState(stateDir, labPath string) (*RunState, error) {
	normalizedLabPath, err := normalizeLabPath(labPath)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(stateDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var latest *RunState
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		state, err := LoadState(filepath.Join(stateDir, entry.Name()))
		if err != nil {
			return nil, err
		}
		stateLabPath, err := normalizeLabPath(state.LabPath)
		if err != nil {
			return nil, err
		}
		if stateLabPath != normalizedLabPath {
			continue
		}
		if latest == nil || state.CreatedAt.After(latest.CreatedAt) {
			latest = state
		}
	}
	return latest, nil
}

func normalizeLabPath(path string) (string, error) {
	absPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return absPath, nil
	}
	return realPath, nil
}
