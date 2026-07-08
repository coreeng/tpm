package builder

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/coreeng/tpm/pkg/module"
	"gopkg.in/yaml.v3"
)

// writeModule marshals a built module to YAML and writes to outdir/module.yaml.
func writeModule(mod *module.BuiltModule, outdir string) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outdir, 0750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(mod)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	// Write to file
	outPath := filepath.Join(outdir, "module.yaml")
	if err := os.WriteFile(outPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
