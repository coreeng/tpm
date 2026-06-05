package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// version is the tpm version. It defaults to "dev" for local builds and is
// overridden at release time via -ldflags "-X .../cmd.version=<tag>".
var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "tpm",
	Short: "Training Platform Module CLI",
	Long: `tpm is a command-line tool for authoring and managing Training Platform modules and labs.

It provides commands to scaffold, list, validate, build, and run training
modules and labs from your local module repository.`,
	Version: version,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(validateChangesCmd)
	rootCmd.AddCommand(generateCodesCmd)
	rootCmd.AddCommand(generateMarkdownCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(labCmd)
}
