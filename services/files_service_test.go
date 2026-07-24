package services

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestFilesServicePlaintext(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("FILES_DIR", dir)
	t.Setenv("FILES_ENCRYPTED", "false")

	svc := NewFilesService()
	if svc.IsEncrypted() {
		t.Fatal("expected plaintext mode")
	}

	rel := "actions/234/note.txt"
	if err := svc.PutFile(rel, bytes.NewReader([]byte("hello"))); err != nil {
		t.Fatal(err)
	}
	disk := filepath.Join(dir, filepath.FromSlash(rel))
	if _, err := os.Stat(disk); err != nil {
		t.Fatalf("expected plaintext file at %s: %v", disk, err)
	}

	data, name, err := svc.GetFile(rel)
	if err != nil {
		t.Fatal(err)
	}
	if name != "note.txt" || string(data) != "hello" {
		t.Fatalf("got name=%q data=%q", name, data)
	}

	list, err := svc.ListFiles()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range list {
		if f.Path == rel && !f.IsDir {
			found = true
		}
	}
	if !found {
		t.Fatalf("list missing %s: %#v", rel, list)
	}
}

func TestFilesServiceEncryptedBinMapping(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("FILES_DIR", dir)
	t.Setenv("FILES_ENCRYPTED", "true")

	svc := NewFilesService()
	if !svc.IsEncrypted() {
		t.Fatal("expected encrypted mode")
	}

	rel := "actions/234/note.txt"
	ciphertext := []byte("RCLONE\x00\x00opaque")
	if err := svc.PutFile(rel, bytes.NewReader(ciphertext)); err != nil {
		t.Fatal(err)
	}

	disk := filepath.Join(dir, filepath.FromSlash(rel+".bin"))
	if _, err := os.Stat(disk); err != nil {
		t.Fatalf("expected .bin on disk at %s: %v", disk, err)
	}
	if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); !os.IsNotExist(err) {
		t.Fatal("plaintext path should not exist when encrypted")
	}

	data, name, err := svc.GetFile(rel)
	if err != nil {
		t.Fatal(err)
	}
	if name != "note.txt" {
		t.Fatalf("logical name should omit .bin, got %q", name)
	}
	if !bytes.Equal(data, ciphertext) {
		t.Fatalf("ciphertext mismatch: %q", data)
	}

	// Accidental .bin in API path still resolves to the same disk file.
	data2, name2, err := svc.GetFile(rel + ".bin")
	if err != nil {
		t.Fatal(err)
	}
	if name2 != "note.txt" || !bytes.Equal(data2, ciphertext) {
		t.Fatalf("suffix strip failed: name=%q data=%q", name2, data2)
	}

	list, err := svc.ListFiles()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range list {
		if f.Path == rel && !f.IsDir {
			found = true
		}
		if !f.IsDir && filepath.Ext(f.Path) == ".bin" {
			t.Fatalf("list should strip .bin, got %q", f.Path)
		}
	}
	if !found {
		t.Fatalf("list missing logical path %s: %#v", rel, list)
	}

	if err := svc.DeleteFile(rel); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(disk); !os.IsNotExist(err) {
		t.Fatal("expected .bin removed")
	}
}

func TestResolvePathTraversal(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("FILES_DIR", dir)
	t.Setenv("FILES_ENCRYPTED", "true")
	svc := NewFilesService()

	if _, _, err := svc.GetFile("../outside.txt"); err != ErrInvalidFilePath {
		t.Fatalf("expected ErrInvalidFilePath, got %v", err)
	}
}
