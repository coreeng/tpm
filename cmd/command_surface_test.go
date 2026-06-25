package cmd

import "testing"

func TestRootCommandRegistersModuleValidationCommands(t *testing.T) {
	for _, name := range []string{
		"list",
		"validate",
		"build",
		"validate-changes",
		"validate-artifact",
		"generate-codes",
		"generate-markdown",
	} {
		if _, _, err := rootCmd.Find([]string{name}); err != nil {
			t.Fatalf("root command missing %q: %v", name, err)
		}
	}
}
