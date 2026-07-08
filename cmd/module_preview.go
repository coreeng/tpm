package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/coreeng/tpm/pkg/builder"
	"github.com/coreeng/tpm/pkg/module"
	"github.com/spf13/cobra"
)

type modulePreviewOptions struct {
	addr  string
	open  bool
	watch bool
}

func newModulePreviewCmd() *cobra.Command {
	opts := &modulePreviewOptions{}
	cmd := &cobra.Command{
		Use:   "preview <module-path>",
		Short: "Preview a full module in a local web UI",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runModulePreview(cmd.Context(), cmd, args[0], opts)
		},
	}
	cmd.Flags().StringVar(&opts.addr, "addr", "127.0.0.1:0", "Address to listen on")
	cmd.Flags().BoolVar(&opts.open, "open", false, "Open the preview URL in the default browser")
	cmd.Flags().BoolVar(&opts.watch, "watch", false, "Reload module metadata and markdown on each browser refresh")
	return cmd
}

func runModulePreview(ctx context.Context, cmd *cobra.Command, modulePath string, opts *modulePreviewOptions) error {
	var loaded *module.BuiltModule
	var err error
	if !opts.watch {
		loaded, err = compilePreviewModule(modulePath)
		if err != nil {
			return err
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		current := loaded
		if opts.watch {
			var err error
			current, err = compilePreviewModule(modulePath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if err := modulePreviewTemplate.Execute(w, current); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	listener, err := net.Listen("tcp", opts.addr)
	if err != nil {
		return err
	}
	defer func() {
		_ = listener.Close()
	}()

	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	errCh := make(chan error, 1)
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	url := "http://" + listener.Addr().String()
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Module preview: %s\n", url); err != nil {
		return err
	}
	if opts.watch {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "watch: reloading module metadata and markdown on refresh"); err != nil {
			return err
		}
	}
	if opts.open {
		_ = openBrowser(url)
	}

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

func compilePreviewModule(modulePath string) (*module.BuiltModule, error) {
	_, _, built, err := builder.Compile(modulePath, "", "")
	return built, err
}

var modulePreviewTemplate = mustPreviewTemplate("module_preview.tmpl")
