package apputils

import (
	"namelesswatch/internal/appconf"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// LocalFileHandler exposes files under the app data directory through /local/*.
type LocalFileHandler struct {
	appDir string
}

func NewLocalFileHandler() (*LocalFileHandler, error) {
	appDir, err := appconf.GetDataDir()
	if err != nil {
		return nil, err
	}
	return &LocalFileHandler{appDir: appDir}, nil
}

func (h *LocalFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/local/") {
		http.NotFound(w, r)
		return
	}

	baseDir, err := filepath.Abs(h.appDir)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	relativePath := strings.TrimPrefix(r.URL.Path, "/local/")
	fullPath := filepath.Join(baseDir, filepath.FromSlash(relativePath))
	cleanPath, err := filepath.Abs(filepath.Clean(fullPath))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if cleanPath != baseDir && !strings.HasPrefix(cleanPath, baseDir+string(os.PathSeparator)) {
		http.NotFound(w, r)
		return
	}

	if info, err := os.Stat(cleanPath); err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeFile(w, r, cleanPath)
}
