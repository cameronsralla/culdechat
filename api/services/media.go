package services

import (
	"bytes"
	"errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	maxProfileURLLen = 2048
	maxPhotoBytes    = 2 << 20
)

func validateProfilePictureURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	if utf8.RuneCountInString(raw) > maxProfileURLLen {
		return "", errors.New("profile_picture_url is too long")
	}
	if strings.HasPrefix(raw, "/api/media/") {
		if strings.Contains(raw, "..") {
			return "", errors.New("invalid profile_picture_url")
		}
		return raw, nil
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return "", errors.New("profile_picture_url must be an http(s) URL or uploaded media path")
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
		return raw, nil
	default:
		return "", errors.New("profile_picture_url must be an http(s) URL")
	}
}

func uploadDir() string {
	if v := strings.TrimSpace(os.Getenv("CULDECHAT_UPLOAD_DIR")); v != "" {
		return v
	}
	return filepath.Join("..", "data", "uploads")
}

func sniffImageExt(head []byte) (string, error) {
	if len(head) >= 3 && head[0] == 0xff && head[1] == 0xd8 && head[2] == 0xff {
		return ".jpg", nil
	}
	if len(head) >= 8 && bytes.Equal(head[:8], []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}) {
		return ".png", nil
	}
	if len(head) >= 12 && bytes.Equal(head[:4], []byte("RIFF")) && bytes.Equal(head[8:12], []byte("WEBP")) {
		return ".webp", nil
	}
	return "", errors.New("photo must be a jpeg, png, or webp image")
}

// SaveProfilePhoto writes a validated image and returns the public API path.
func SaveProfilePhoto(userID uuid.UUID, r io.Reader) (string, error) {
	limited := io.LimitReader(r, maxPhotoBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return "", err
	}
	if len(data) == 0 || int64(len(data)) > maxPhotoBytes {
		return "", errors.New("photo must be between 1 byte and 2MB")
	}
	ext, err := sniffImageExt(data)
	if err != nil {
		return "", err
	}
	dir := filepath.Join(uploadDir(), "profile")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", err
	}
	name := userID.String() + ext
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0o640); err != nil {
		return "", err
	}
	return "/api/media/profile/" + name, nil
}

// ResolveProfilePhotoPath maps a URL filename to a file under the upload dir.
func ResolveProfilePhotoPath(name string) (string, error) {
	base := filepath.Base(name)
	if base != name || strings.Contains(base, "..") {
		return "", errors.New("invalid filename")
	}
	ext := strings.ToLower(filepath.Ext(base))
	switch ext {
	case ".jpg", ".png", ".webp":
	default:
		return "", errors.New("invalid filename")
	}
	full := filepath.Join(uploadDir(), "profile", base)
	if _, err := os.Stat(full); err != nil {
		return "", ErrNotFound
	}
	return full, nil
}
