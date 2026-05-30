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

	info, err := os.Stat(cleanPath)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	if contentType := localFileContentType(cleanPath); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	if isLocalAudioFile(cleanPath) {
		w.Header().Set("Accept-Ranges", "bytes")
	}

	file, err := os.Open(cleanPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	http.ServeContent(w, r, info.Name(), info.ModTime(), file)
}

func localFileContentType(filePath string) string {
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	case ".bmp":
		return "image/bmp"
	case ".mp3":
		return "audio/mpeg"
	case ".ogg":
		return "audio/ogg"
	case ".wav":
		return "audio/wav"
	case ".m4a":
		return "audio/mp4"
	case ".webm":
		return "audio/webm"
	default:
		return ""
	}
}

func isLocalAudioFile(filePath string) bool {
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".mp3", ".ogg", ".wav", ".m4a", ".webm":
		return true
	default:
		return false
	}
}
