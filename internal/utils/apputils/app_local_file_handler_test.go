package apputils

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalFileHandlerServesAudioWithContentTypeAndRange(t *testing.T) {
	root := t.TempDir()
	audioPath := filepath.Join(root, "story-assets", "game-1", "bgm", "theme.mp3")
	if err := os.MkdirAll(filepath.Dir(audioPath), 0o755); err != nil {
		t.Fatalf("create audio dir: %v", err)
	}
	if err := os.WriteFile(audioPath, []byte("0123456789"), 0o600); err != nil {
		t.Fatalf("write audio: %v", err)
	}

	handler := &LocalFileHandler{appDir: root}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/local/story-assets/game-1/bgm/theme.mp3", nil)
	handler.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.StatusCode)
	}
	if got := response.Header.Get("Content-Type"); got != "audio/mpeg" {
		t.Fatalf("expected audio/mpeg content type, got %q", got)
	}
	if got := response.Header.Get("Accept-Ranges"); got != "bytes" {
		t.Fatalf("expected byte range support, got %q", got)
	}

	recorder = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/local/story-assets/game-1/bgm/theme.mp3", nil)
	request.Header.Set("Range", "bytes=2-5")
	handler.ServeHTTP(recorder, request)

	response = recorder.Result()
	defer response.Body.Close()
	if response.StatusCode != http.StatusPartialContent {
		t.Fatalf("expected 206, got %d", response.StatusCode)
	}
	if got := response.Header.Get("Content-Range"); got != "bytes 2-5/10" {
		t.Fatalf("expected content range bytes 2-5/10, got %q", got)
	}
	if got := recorder.Body.String(); got != "2345" {
		t.Fatalf("expected ranged body, got %q", got)
	}
}

func TestLocalFileHandlerKeepsImageContentType(t *testing.T) {
	root := t.TempDir()
	imagePath := filepath.Join(root, "story-assets", "game-1", "photo", "scene.png")
	if err := os.MkdirAll(filepath.Dir(imagePath), 0o755); err != nil {
		t.Fatalf("create image dir: %v", err)
	}
	if err := os.WriteFile(imagePath, []byte("png"), 0o600); err != nil {
		t.Fatalf("write image: %v", err)
	}

	handler := &LocalFileHandler{appDir: root}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/local/story-assets/game-1/photo/scene.png", nil)
	handler.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.StatusCode)
	}
	if got := response.Header.Get("Content-Type"); got != "image/png" {
		t.Fatalf("expected image/png content type, got %q", got)
	}
}
