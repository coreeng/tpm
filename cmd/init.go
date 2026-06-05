package cmd

import (
	"fmt"
	"path/filepath"

	moduleinit "github.com/coreeng/tpm/pkg/init"
	"github.com/coreeng/tpm/pkg/lab"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create module and lab skeletons",
	Long: `Create starter files for Training Platform modules and local labs.

Use standalone labs when you want a local lab outside a module.
Use module-backed labs when you want module metadata and lab runtime files together.`,
}

var initModuleCmd = &cobra.Command{
	Use:   "module <name>",
	Short: "Create a module skeleton",
	Args:  cobra.ExactArgs(1),
	RunE:  runInitModule,
}

var initLabCmd = &cobra.Command{
	Use:   "lab <path>",
	Short: "Create a Pod image lab skeleton",
	Long: `Create a Pod image lab skeleton.

Standalone lab:
  tpm init lab <path>

Module-backed lab:
  tpm init lab <module> <chapter> <lab-name> --module-backed`,
	RunE: runInitLab,
}

func init() {
	initLabCmd.Flags().Bool("module-backed", false, "Create module metadata and lab runtime files under an existing module root")
	initLabCmd.Flags().String("artifact-registry", lab.DefaultArtifactRegistry, "OCI registry for module-backed starter and validator artifact URIs")
	initCmd.AddCommand(initModuleCmd)
	initCmd.AddCommand(initLabCmd)
}

func runInitModule(cmd *cobra.Command, args []string) error {
	target := filepath.Clean(args[0])
	name := filepath.Base(target)
	dir := filepath.Dir(target)

	if err := moduleinit.ScaffoldModuleSkeleton(dir, moduleinit.ModuleScaffoldOptions{Name: name}); err != nil {
		return err
	}

	_, err := fmt.Fprintf(cmd.OutOrStdout(), `Created module %s.

Next steps:
  cd %s
  tpm init lab . <chapter> <lab-name> --module-backed
  tpm validate
`, target, target)
	return err
}

func runInitLab(cmd *cobra.Command, args []string) error {
	moduleBacked, err := cmd.Flags().GetBool("module-backed")
	if err != nil {
		return err
	}
	defer cmd.Flags().Set("module-backed", "false")

	if moduleBacked {
		return runInitModuleBackedLab(cmd, args)
	}
	return runInitStandaloneLab(cmd, args)
}

func runInitStandaloneLab(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("standalone lab requires exactly 1 argument: tpm init lab <path>")
	}

	target := filepath.Clean(args[0])
	if err := lab.ScaffoldStandalone(target, lab.ScaffoldOptions{Name: filepath.Base(target)}); err != nil {
		return err
	}

	_, err := fmt.Fprintf(cmd.OutOrStdout(), `Created standalone lab %s.

Next steps:
  tpm lab run %s
`, target, target)
	return err
}

func runInitModuleBackedLab(cmd *cobra.Command, args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("module-backed lab requires exactly 3 arguments: tpm init lab <module> <chapter> <lab-name> --module-backed")
	}

	moduleDir := filepath.Clean(args[0])
	chapter := args[1]
	labName := args[2]
	artifactRegistry, err := cmd.Flags().GetString("artifact-registry")
	if err != nil {
		return err
	}
	if err := lab.ScaffoldModuleBacked(moduleDir, lab.ModuleBackedScaffoldOptions{Chapter: chapter, Name: labName, ArtifactRegistry: artifactRegistry}); err != nil {
		return err
	}
	runtimeDir := filepath.Join(moduleDir, "assessments", chapter, labName)

	_, err = fmt.Fprintf(cmd.OutOrStdout(), `Created module-backed lab %s.

Next steps:
  tpm lab run %s
  tpm validate
`, filepath.Join(moduleDir, "module", chapter, "assessments", labName), runtimeDir)
	return err
}
