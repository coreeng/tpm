package cmd

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPreviewEventsReportFileChanges(t *testing.T) {
	oldInterval := previewWatchInterval
	previewWatchInterval = 20 * time.Millisecond
	t.Cleanup(func() {
		previewWatchInterval = oldInterval
	})

	root := t.TempDir()
	sourcePath := filepath.Join(root, "module.yaml")
	if err := os.WriteFile(sourcePath, []byte("title: Old\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	if err := registerPreviewHandlers(mux, func() (any, error) {
		return map[string]string{"kind": "test"}, nil
	}, []string{root}); err != nil {
		t.Fatalf("registerPreviewHandlers returned error: %v", err)
	}

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(server.URL + "/api/events")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	reader := bufio.NewReader(resp.Body)
	readPreviewEvent(t, reader, "event: ready")

	if err := os.WriteFile(sourcePath, []byte("title: Updated module title\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	readPreviewEvent(t, reader, "event: preview-changed")
}

func readPreviewEvent(t *testing.T, reader *bufio.Reader, want string) {
	t.Helper()

	deadline := time.After(2 * time.Second)
	for {
		lineCh := make(chan string, 1)
		errCh := make(chan error, 1)
		go func() {
			line, err := reader.ReadString('\n')
			if err != nil {
				errCh <- err
				return
			}
			lineCh <- line
		}()

		select {
		case <-deadline:
			t.Fatalf("timed out waiting for %q", want)
		case err := <-errCh:
			t.Fatalf("read event stream: %v", err)
		case line := <-lineCh:
			if strings.TrimSpace(line) == want {
				return
			}
		}
	}
}
