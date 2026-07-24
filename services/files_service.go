package services

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const defaultFilesDir = "/Users/arturgudiev/Data/dashboard_files"

var (
	ErrInvalidFilePath = errors.New("invalid file path")
	ErrFileNotFound    = errors.New("file not found")
	ErrNotAFile        = errors.New("path is not a file")
)

// FileInfo describes a file under the configured files directory.
type FileInfo struct {
	Path  string `json:"path"`
	Size  int64  `json:"size"`
	IsDir bool   `json:"isDir"`
}

// FilesConfig is returned to clients so they know storage mode.
type FilesConfig struct {
	Encrypted bool   `json:"encrypted"`
	BaseDir   string `json:"baseDir"`
}

// FilesService serves files from a host directory.
// When Encrypted is true, stored bytes are treated as opaque ciphertext (TNO);
// this service does not encrypt or decrypt.
type FilesService struct {
	baseDir   string
	encrypted bool
}

func NewFilesService() *FilesService {
	baseDir := os.Getenv("FILES_DIR")
	if baseDir == "" {
		baseDir = defaultFilesDir
	}
	baseDir = filepath.Clean(baseDir)

	encrypted := strings.EqualFold(os.Getenv("FILES_ENCRYPTED"), "true") ||
		os.Getenv("FILES_ENCRYPTED") == "1"

	return &FilesService{
		baseDir:   baseDir,
		encrypted: encrypted,
	}
}

func (s *FilesService) Config() FilesConfig {
	return FilesConfig{
		Encrypted: s.encrypted,
		BaseDir:   s.baseDir,
	}
}

func (s *FilesService) IsEncrypted() bool {
	return s.encrypted
}

func (s *FilesService) BaseDir() string {
	return s.baseDir
}

const encryptedSuffix = ".bin"

// resolvePath maps a logical relative path to an absolute path under baseDir.
// When Encrypted is true, disk paths use a trailing .bin (rclone crypt suffix);
// the logical API path never includes .bin.
func (s *FilesService) resolvePath(relPath string) (string, error) {
	relPath = strings.TrimSpace(relPath)
	relPath = strings.TrimPrefix(relPath, "/")
	if relPath == "" || relPath == "." {
		return "", ErrInvalidFilePath
	}
	if strings.Contains(relPath, "\x00") {
		return "", ErrInvalidFilePath
	}

	// Clients always use logical paths; strip a mistaken .bin suffix.
	if s.encrypted && strings.HasSuffix(relPath, encryptedSuffix) {
		relPath = strings.TrimSuffix(relPath, encryptedSuffix)
	}

	cleaned := filepath.Clean(relPath)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(os.PathSeparator)) {
		return "", ErrInvalidFilePath
	}
	if filepath.IsAbs(cleaned) {
		return "", ErrInvalidFilePath
	}

	full := filepath.Join(s.baseDir, cleaned)
	full = filepath.Clean(full)

	baseWithSep := s.baseDir + string(os.PathSeparator)
	if full != s.baseDir && !strings.HasPrefix(full, baseWithSep) {
		return "", ErrInvalidFilePath
	}

	if s.encrypted {
		full += encryptedSuffix
		// Re-check containment after appending suffix (suffix cannot escape, but be explicit).
		if !strings.HasPrefix(full, baseWithSep) {
			return "", ErrInvalidFilePath
		}
	}

	return full, nil
}

func logicalRelPath(rel string, isDir bool, encrypted bool) string {
	if !encrypted || isDir {
		return rel
	}
	if strings.HasSuffix(rel, encryptedSuffix) {
		return strings.TrimSuffix(rel, encryptedSuffix)
	}
	return rel
}

// GetFile reads a file by logical relative path and returns its contents.
// When Encrypted, reads the on-disk .bin file and returns opaque ciphertext.
func (s *FilesService) GetFile(relPath string) ([]byte, string, error) {
	full, err := s.resolvePath(relPath)
	if err != nil {
		return nil, "", err
	}

	info, err := os.Stat(full)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", ErrFileNotFound
		}
		return nil, "", err
	}
	if info.IsDir() {
		return nil, "", ErrNotAFile
	}

	data, err := os.ReadFile(full)
	if err != nil {
		return nil, "", err
	}

	name := filepath.Base(full)
	if s.encrypted {
		name = strings.TrimSuffix(name, encryptedSuffix)
	}
	return data, name, nil
}

// ListFiles walks the files directory and returns logical relative paths.
// When Encrypted, strips the trailing .bin suffix from file paths.
func (s *FilesService) ListFiles() ([]FileInfo, error) {
	info, err := os.Stat(s.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []FileInfo{}, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("FILES_DIR is not a directory: %s", s.baseDir)
	}

	var files []FileInfo
	err = filepath.WalkDir(s.baseDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == s.baseDir {
			return nil
		}
		rel, err := filepath.Rel(s.baseDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		fi, err := d.Info()
		if err != nil {
			return err
		}
		files = append(files, FileInfo{
			Path:  logicalRelPath(rel, d.IsDir(), s.encrypted),
			Size:  fi.Size(),
			IsDir: d.IsDir(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if files == nil {
		files = []FileInfo{}
	}
	return files, nil
}

// PutFile writes content to a logical relative path (creates parent dirs as needed).
// When Encrypted, writes to the corresponding .bin path; body should already be ciphertext.
func (s *FilesService) PutFile(relPath string, r io.Reader) error {
	full, err := s.resolvePath(relPath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}

	f, err := os.OpenFile(full, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	return err
}

// DeleteFile removes a file by logical relative path.
func (s *FilesService) DeleteFile(relPath string) error {
	full, err := s.resolvePath(relPath)
	if err != nil {
		return err
	}

	info, err := os.Stat(full)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		return err
	}
	if info.IsDir() {
		return ErrNotAFile
	}

	return os.Remove(full)
}
