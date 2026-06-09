package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/coreeng/tpm/pkg/lab"
	"github.com/spf13/cobra"
)

var labCmd = newLabCmd()

func newLabCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "lab",
		Short:        "Run and inspect local labs",
		Long:         "Run, inspect, and clean up local labs.",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newLabRunCmd())
	cmd.AddCommand(newLabListCmd())
	cmd.AddCommand(newLabStatusCmd())
	cmd.AddCommand(newLabCleanupCmd())
	return cmd
}

func newLabListCmd() *cobra.Command {
	opts := lab.Options{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List active local labs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := lab.List(cmd.Context(), opts)
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), output)
			return err
		},
	}
	cmd.Flags().StringVar(&opts.StateDir, "state-dir", "", "lab state directory (default "+lab.DefaultStateDirHelp+")")
	cmd.Flags().BoolVar(&opts.AllowNonKind, "allow-non-kind", false, "allow listing labs against a kubectl context that is not kind-")
	return cmd
}

func newLabRunCmd() *cobra.Command {
	opts := lab.Options{}
	cmd := &cobra.Command{
		Use:   "run <lab-path>",
		Short: "Run a local lab",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			chartDir := strings.TrimSpace(opts.ChartDir)
			chartURI := strings.TrimSpace(opts.ChartURI)
			chartURIChanged := cmd.Flags().Changed("chart-uri")
			if chartDir != "" && chartURI != "" {
				return fmt.Errorf("set either chart-dir or chart-uri, not both")
			}
			if chartURIChanged && chartURI == "" {
				return fmt.Errorf("chart-uri must not be blank")
			}
			if chartDir == "" && chartURI == "" {
				return fmt.Errorf("chart-dir or chart-uri must be set")
			}
			if chartURI != "" && strings.TrimSpace(opts.ChartVersion) == "" {
				return fmt.Errorf("chart-version must not be blank")
			}
			opts.ChartDir = chartDir
			opts.ChartURI = chartURI
			opts.LabPath = args[0]
			opts.LogWriter = cmd.OutOrStdout()
			state, err := lab.Run(cmd.Context(), opts)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Lab %s running\nSystem namespace: %s\nWorkspace namespace: %s\nRegistry URL: %s\nRegistry username: %s\nRegistry token: %s\n", state.RunID, state.SystemNamespace, state.WorkspaceNamespace, state.RegistryURL, state.RegistryUsername, state.RegistryToken)
			return err
		},
	}
	addSharedLabFlags(cmd, &opts)
	cmd.Flags().StringVar(&opts.ChartDir, "chart-dir", "", "Path to a local lab runtime Helm chart directory")
	cmd.Flags().StringVar(&opts.ChartURI, "chart-uri", "", "OCI URI for the lab runtime Helm chart")
	cmd.Flags().StringVar(&opts.ChartVersion, "chart-version", "", "Version of the lab runtime Helm chart")
	cmd.Flags().StringVar(&opts.ValidatorRegistry, "validator-registry", lab.DefaultArtifactRegistry, "Registry for the locally built validator image")
	cmd.Flags().StringVar(&opts.RegistryDomain, "registry-domain", lab.DefaultArtifactRegistry, "Learner registry domain passed to the lab runtime chart")
	cmd.Flags().BoolVar(&opts.AssumeImageAccessible, "assume-image-accessible", false, "assume a non-kind cluster can pull the local validator image tag")
	cmd.Flags().DurationVar(&opts.CheckInterval, "check-interval", 5*time.Second, "validator check interval")
	return cmd
}

func newLabStatusCmd() *cobra.Command {
	opts := lab.Options{}
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Print local lab status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := lab.Status(cmd.Context(), opts)
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), output)
			return err
		},
	}
	cmd.Flags().StringVar(&opts.ID, "id", "", "lab run ID")
	cmd.Flags().StringVar(&opts.StateDir, "state-dir", "", "lab state directory (default "+lab.DefaultStateDirHelp+")")
	return cmd
}

func newLabCleanupCmd() *cobra.Command {
	opts := lab.Options{}
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up a local lab run",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return lab.Cleanup(cmd.Context(), opts)
		},
	}
	cmd.Flags().StringVar(&opts.ID, "id", "", "lab run ID")
	cmd.Flags().StringVar(&opts.StateDir, "state-dir", "", "lab state directory (default "+lab.DefaultStateDirHelp+")")
	cmd.Flags().BoolVar(&opts.AllowNonKind, "allow-non-kind", false, "allow running against a kubectl context that is not kind-")
	return cmd
}

func addSharedLabFlags(cmd *cobra.Command, opts *lab.Options) {
	cmd.Flags().StringVar(&opts.ID, "id", "", "lab run ID (default generated)")
	cmd.Flags().StringVar(&opts.StateDir, "state-dir", "", "lab state directory (default "+lab.DefaultStateDirHelp+")")
	cmd.Flags().BoolVar(&opts.AllowNonKind, "allow-non-kind", false, "allow running against a kubectl context that is not kind-")
}
