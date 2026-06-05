package cmd

import (
	"fmt"
	"os"

	"github.com/coreeng/tpm/pkg/module"
	"github.com/coreeng/tpm/pkg/pathutil"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all training platform modules",
	Long: `List all directories in the repository root that contain training platform modules.

A directory is considered a module if it contains a module/module.yaml file.`,
	Run: runList,
}

func runList(cmd *cobra.Command, args []string) {
	// Get repository root using shared utility
	rootDir, err := pathutil.GetRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Find all modules
	modules, err := module.FindModules(rootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding modules: %v\n", err)
		os.Exit(1)
	}

	// Print module names (one per line for easy scripting)
	for _, mod := range modules {
		fmt.Println(mod)
	}
}
