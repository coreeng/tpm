package cmd

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/coreeng/tpm/internal/previewui"
)

var previewWatchInterval = time.Second

func registerPreviewHandlers(mux *http.ServeMux, load func() (any, error), watchRoots []string) error {
	appFS, err := previewui.FS()
	if err != nil {
		return fmt.Errorf("load preview UI assets: %w", err)
	}
	indexHTML, err := fs.ReadFile(appFS, "index.html")
	if err != nil {
		return fmt.Errorf("load preview UI index: %w", err)
	}

	mux.Handle("/assets/", http.FileServer(http.FS(appFS)))
	mux.HandleFunc("/api/preview", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		data, err := load()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		streamPreviewEvents(w, r, watchRoots)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		if r.Method == http.MethodHead {
			return
		}
		_, _ = w.Write(indexHTML)
	})
	return nil
}

func streamPreviewEvents(w http.ResponseWriter, r *http.Request, watchRoots []string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	watchRoots = cleanPreviewWatchRoots(watchRoots)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Connection", "keep-alive")

	lastSnapshot, _ := snapshotPreviewFiles(watchRoots)
	_, _ = fmt.Fprint(w, "event: ready\ndata: {}\n\n")
	flusher.Flush()

	ticker := time.NewTicker(previewWatchInterval)
	defer ticker.Stop()

	keepAlive := time.NewTicker(15 * time.Second)
	defer keepAlive.Stop()

	version := 0
	for {
		select {
		case <-r.Context().Done():
			return
		case <-keepAlive.C:
			_, _ = fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		case <-ticker.C:
			if len(watchRoots) == 0 {
				continue
			}
			nextSnapshot, err := snapshotPreviewFiles(watchRoots)
			if err != nil {
				_, _ = fmt.Fprintf(w, "event: preview-error\ndata: %s\n\n", mustJSON(map[string]string{"error": err.Error()}))
				flusher.Flush()
				continue
			}
			if nextSnapshot == lastSnapshot {
				continue
			}
			lastSnapshot = nextSnapshot
			version++
			_, _ = fmt.Fprintf(w, "event: preview-changed\ndata: %s\n\n", mustJSON(map[string]int{"version": version}))
			flusher.Flush()
		}
	}
}

func cleanPreviewWatchRoots(roots []string) []string {
	seen := make(map[string]struct{}, len(roots))
	cleaned := make([]string, 0, len(roots))
	for _, root := range roots {
		root = strings.TrimSpace(root)
		if root == "" {
			continue
		}
		abs, err := filepath.Abs(filepath.Clean(root))
		if err == nil {
			root = abs
		} else {
			root = filepath.Clean(root)
		}
		if _, ok := seen[root]; ok {
			continue
		}
		seen[root] = struct{}{}
		cleaned = append(cleaned, root)
	}
	return cleaned
}

func snapshotPreviewFiles(roots []string) (uint64, error) {
	hash := fnv.New64a()
	for _, root := range cleanPreviewWatchRoots(roots) {
		info, err := os.Stat(root)
		if err != nil {
			_, _ = fmt.Fprintf(hash, "missing\x00%s\x00%s\n", root, err)
			continue
		}
		if !info.IsDir() {
			writePreviewFileSnapshot(hash, root, filepath.Base(root), info)
			continue
		}
		if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				_, _ = fmt.Fprintf(hash, "walk-error\x00%s\x00%s\n", path, walkErr)
				return nil
			}
			if entry.IsDir() {
				if shouldSkipPreviewWatchDir(entry.Name()) {
					return filepath.SkipDir
				}
				return nil
			}
			info, err := entry.Info()
			if err != nil {
				_, _ = fmt.Fprintf(hash, "stat-error\x00%s\x00%s\n", path, err)
				return nil
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				rel = path
			}
			writePreviewFileSnapshot(hash, root, rel, info)
			return nil
		}); err != nil {
			return 0, err
		}
	}
	return hash.Sum64(), nil
}

func shouldSkipPreviewWatchDir(name string) bool {
	switch name {
	case ".git", ".tpm", "node_modules", "dist", "build", ".build":
		return true
	default:
		return false
	}
}

func writePreviewFileSnapshot(hash interface {
	Write([]byte) (int, error)
}, root string, rel string, info os.FileInfo) {
	_, _ = fmt.Fprintf(hash, "file\x00%s\x00%s\x00%d\x00%d\n", root, rel, info.Size(), info.ModTime().UnixNano())
}

func mustJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(data)
}
