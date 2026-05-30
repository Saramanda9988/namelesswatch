package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"namelesswatch/internal/appconf"
	"namelesswatch/internal/roleplay"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const storyAssetsDirName = "story-assets"

func materializeLibraryGameAssets(game roleplay.LibraryGame) (roleplay.LibraryGame, error) {
	if game.ID == "" || len(game.Files) == 0 {
		return game, nil
	}

	nextFiles := make(map[string]string, len(game.Files))
	for name, content := range game.Files {
		nextFiles[name] = content
	}

	for name, content := range game.Files {
		if !isPackAssetPath(name) || strings.HasPrefix(strings.TrimSpace(content), "/local/") {
			continue
		}

		data, err := assetBytes(content)
		if err != nil {
			return roleplay.LibraryGame{}, fmt.Errorf("read imported asset %s: %w", name, err)
		}
		if len(data) == 0 {
			continue
		}

		assetPath, assetURL, err := storyAssetPathAndURL(game.ID, name)
		if err != nil {
			return roleplay.LibraryGame{}, err
		}
		if err := os.MkdirAll(filepath.Dir(assetPath), 0o755); err != nil {
			return roleplay.LibraryGame{}, fmt.Errorf("create asset directory: %w", err)
		}
		if err := os.WriteFile(assetPath, data, 0o600); err != nil {
			return roleplay.LibraryGame{}, fmt.Errorf("write imported asset %s: %w", name, err)
		}

		nextFiles[name] = assetURL
	}

	game.Files = nextFiles
	return game, nil
}

func deleteLibraryGameAssets(gameID string) error {
	if strings.TrimSpace(gameID) == "" {
		return nil
	}
	root, err := appconf.GetSubDir(storyAssetsDirName)
	if err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(root, safePathSegment(gameID)))
}

func assetBytes(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	if strings.HasPrefix(value, "data:") {
		comma := strings.Index(value, ",")
		if comma < 0 {
			return nil, nil
		}
		header := value[:comma]
		payload := value[comma+1:]
		if strings.Contains(header, ";base64") {
			data, err := base64.StdEncoding.DecodeString(payload)
			if err != nil {
				return nil, nil
			}
			return data, nil
		}
		decoded, err := url.PathUnescape(payload)
		if err != nil {
			return nil, nil
		}
		return []byte(decoded), nil
	}

	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "blob:") {
		return nil, nil
	}
	if strings.HasPrefix(value, "file://") {
		parsed, err := url.Parse(value)
		if err != nil {
			return nil, err
		}
		value = parsed.Path
	}

	file, err := os.Open(value)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}

func storyAssetPathAndURL(gameID, relativeName string) (string, string, error) {
	root, err := appconf.GetSubDir(storyAssetsDirName)
	if err != nil {
		return "", "", err
	}

	relativeName = strings.TrimPrefix(strings.ReplaceAll(relativeName, "\\", "/"), "/")
	cleanName := path.Clean(relativeName)
	if cleanName == "." || strings.HasPrefix(cleanName, "../") || cleanName == ".." {
		return "", "", fmt.Errorf("invalid asset path %q", relativeName)
	}

	segments := append([]string{safePathSegment(gameID)}, strings.Split(cleanName, "/")...)
	localPath := filepath.Join(append([]string{root}, segments...)...)
	return localPath, "/local/" + encodePathSegments(append([]string{storyAssetsDirName}, segments...)), nil
}

func isPackImagePath(name string) bool {
	lower := strings.ToLower(strings.ReplaceAll(name, "\\", "/"))
	if !strings.HasPrefix(lower, "photo/") && !strings.HasPrefix(lower, "map/") {
		return false
	}
	switch filepath.Ext(lower) {
	case ".png", ".jpg", ".jpeg", ".webp", ".gif", ".bmp":
		return true
	default:
		return false
	}
}

func isPackAudioPath(name string) bool {
	lower := strings.ToLower(strings.ReplaceAll(name, "\\", "/"))
	if !strings.HasPrefix(lower, "bgm/") {
		return false
	}
	switch filepath.Ext(lower) {
	case ".mp3", ".ogg", ".wav", ".m4a", ".webm":
		return true
	default:
		return false
	}
}

func isPackAssetPath(name string) bool {
	return isPackImagePath(name) || isPackAudioPath(name)
}

func encodePathSegments(segments []string) string {
	encoded := make([]string, 0, len(segments))
	for _, segment := range segments {
		encoded = append(encoded, url.PathEscape(segment))
	}
	return strings.Join(encoded, "/")
}

func safePathSegment(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	value = replacer.Replace(value)
	if value == "." || value == ".." {
		return "unknown"
	}
	return value
}
