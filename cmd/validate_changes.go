package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/coreeng/tpm/pkg/git"
	"github.com/spf13/cobra"
)

var validateChangesCmd = &cobra.Command{
	Use:   "validate-changes <old-ref> <new-ref>",
	Short: "Validate that no codes have been removed between git refs",
	Long: `Validate that no codes (UUIDs and semantic identifiers) have been removed
between two git references.

This command ensures backwards compatibility by verifying that all codes present
in the old ref still exist in the new ref. This prevents breaking changes that
would disrupt learner progress tracking.

The command:
  - Collects all codes from modules at the old ref
  - Collects all codes from modules at the new ref
  - Compares and reports any codes that disappeared
  - Allows adding new codes and reordering (non-breaking changes)

Git refs can be:
  - Branch names (main, develop, feature/my-branch)
  - Commit SHAs (abc123, HEAD, HEAD~1)
  - Tags (v1.0.0, release-2024)

This command NEVER modifies the working tree. It uses 'git show' to read files
from different refs without checking them out.

Returns exit code 0 if no codes were removed, 1 if codes are missing.

Examples:
  tpm validate-changes main HEAD          # Compare main branch to current commit
  tpm validate-changes v1.0.0 v1.1.0      # Compare two tagged releases
  tpm validate-changes HEAD~5 HEAD        # Compare last 5 commits
  tpm validate-changes abc123 def456      # Compare two commits by SHA`,
	Args: cobra.ExactArgs(2),
	Run:  runValidateChanges,
}

func runValidateChanges(cmd *cobra.Command, args []string) {
	oldRef := args[0]
	newRef := args[1]

	// Check if we're in a git repository
	if !git.IsGitRepo() {
		fmt.Fprintf(os.Stderr, "Error: Not in a git repository\n")
		os.Exit(1)
	}

	// Validate both refs exist
	fmt.Printf("Validating git references...\n")
	if err := git.ValidateRef(oldRef); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := git.ValidateRef(newRef); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  ✓ Both refs are valid\n\n")

	// Collect codes from old ref
	fmt.Printf("Collecting codes from '%s'...\n", oldRef)
	oldCodes, err := git.CollectCodesAtRef(oldRef)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error collecting codes from %s: %v\n", oldRef, err)
		os.Exit(1)
	}
	fmt.Printf("  ✓ Found %d code(s)\n\n", len(oldCodes))

	// Collect codes from new ref
	fmt.Printf("Collecting codes from '%s'...\n", newRef)
	newCodes, err := git.CollectCodesAtRef(newRef)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error collecting codes from %s: %v\n", newRef, err)
		os.Exit(1)
	}
	fmt.Printf("  ✓ Found %d code(s)\n\n", len(newCodes))

	// Find codes that exist in old but not in new (removed codes)
	removedCodes := make([]git.CodeInfo, 0)
	for code, info := range oldCodes {
		if _, exists := newCodes[code]; !exists {
			removedCodes = append(removedCodes, info)
		}
	}

	// Find new codes (added in new ref)
	addedCodes := make([]git.CodeInfo, 0)
	for code, info := range newCodes {
		if _, exists := oldCodes[code]; !exists {
			addedCodes = append(addedCodes, info)
		}
	}

	// Find codes that moved between parents (breaking change)
	movedCodes := make([]struct {
		Code       string
		OldParent  string
		NewParent  string
		EntityType string
		OldFile    string
		NewFile    string
	}, 0)
	for code, newInfo := range newCodes {
		if oldInfo, exists := oldCodes[code]; exists {
			// Check if parent changed
			if oldInfo.ParentCode != newInfo.ParentCode {
				movedCodes = append(movedCodes, struct {
					Code       string
					OldParent  string
					NewParent  string
					EntityType string
					OldFile    string
					NewFile    string
				}{
					Code:       code,
					OldParent:  oldInfo.ParentCode,
					NewParent:  newInfo.ParentCode,
					EntityType: oldInfo.EntityType,
					OldFile:    oldInfo.FilePath,
					NewFile:    newInfo.FilePath,
				})
			}
		}
	}

	// Print summary
	fmt.Printf("Change summary:\n")
	fmt.Printf("  Codes in %s: %d\n", oldRef, len(oldCodes))
	fmt.Printf("  Codes in %s: %d\n", newRef, len(newCodes))
	fmt.Printf("  Added: %d\n", len(addedCodes))
	fmt.Printf("  Removed: %d\n", len(removedCodes))
	fmt.Printf("  Moved between parents: %d\n", len(movedCodes))
	fmt.Println()

	// Report added codes (informational, not an error)
	if len(addedCodes) > 0 {
		fmt.Printf("✓ %d new code(s) added:\n", len(addedCodes))

		// Sort by code for consistent output
		sort.Slice(addedCodes, func(i, j int) bool {
			return addedCodes[i].Code < addedCodes[j].Code
		})

		// Group by module for better readability
		moduleGroups := groupCodesByModule(addedCodes)
		for _, moduleName := range sortedKeys(moduleGroups) {
			fmt.Printf("  [%s]:\n", moduleName)
			for _, info := range moduleGroups[moduleName] {
				fmt.Printf("    + %s (%s): %s\n", info.Code, info.EntityType, info.FilePath)
			}
		}
		fmt.Println()
	}

	// Report codes that moved between parents (this is an error)
	if len(movedCodes) > 0 {
		fmt.Printf("✗ %d code(s) MOVED between parents (breaking change):\n", len(movedCodes))

		// Sort by code for consistent output
		sort.Slice(movedCodes, func(i, j int) bool {
			return movedCodes[i].Code < movedCodes[j].Code
		})

		for _, moved := range movedCodes {
			fmt.Printf("  Code '%s' (%s):\n", moved.Code, moved.EntityType)
			fmt.Printf("    Old parent: %s (in %s)\n", moved.OldParent, moved.OldFile)
			fmt.Printf("    New parent: %s (in %s)\n", moved.NewParent, moved.NewFile)
		}
		fmt.Println()
	}

	// Report removed codes (this is an error)
	if len(removedCodes) > 0 {
		fmt.Printf("✗ %d code(s) were REMOVED (breaking change):\n", len(removedCodes))

		// Sort by code for consistent output
		sort.Slice(removedCodes, func(i, j int) bool {
			return removedCodes[i].Code < removedCodes[j].Code
		})

		// Group by module for better readability
		moduleGroups := groupCodesByModule(removedCodes)
		for _, moduleName := range sortedKeys(moduleGroups) {
			fmt.Printf("  [%s]:\n", moduleName)
			for _, info := range moduleGroups[moduleName] {
				fmt.Printf("    - %s (%s): %s\n", info.Code, info.EntityType, info.FilePath)
			}
		}
		fmt.Println()
	}

	// Check if any breaking changes occurred
	if len(removedCodes) > 0 || len(movedCodes) > 0 {
		fmt.Println("✗ Validation failed: Breaking changes detected")
		fmt.Println()
		if len(removedCodes) > 0 {
			fmt.Println("Removing codes is a breaking change that can disrupt learner progress.")
		}
		if len(movedCodes) > 0 {
			fmt.Println("Moving codes between parents is a breaking change that can disrupt learner progress.")
			fmt.Println("Content must remain under the same parent entity to maintain learner tracking.")
		}
		fmt.Println()
		fmt.Println("If you need to restructure content, consider:")
		fmt.Println("  1. Deprecating it first (mark as deprecated but keep the structure)")
		fmt.Println("  2. Using a major version bump if restructuring is necessary")
		fmt.Println("  3. Coordinating with the training platform team")
		os.Exit(1)
	}

	// Success
	fmt.Println("✓ No codes were removed - backwards compatibility maintained!")
	fmt.Println("✓ No codes moved between parents - structure preserved!")
	if len(addedCodes) > 0 {
		fmt.Println("✓ New codes were added - this is a non-breaking change")
	}
}

// groupCodesByModule groups codes by their module name
func groupCodesByModule(codes []git.CodeInfo) map[string][]git.CodeInfo {
	groups := make(map[string][]git.CodeInfo)
	for _, code := range codes {
		groups[code.ModuleName] = append(groups[code.ModuleName], code)
	}
	return groups
}

// sortedKeys returns sorted keys from a map
func sortedKeys(m map[string][]git.CodeInfo) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
