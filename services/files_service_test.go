package services

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"arturgudiev/dashboard/ent/schema"
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

func TestListFilesInAbsoluteDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("FILES_DIR", dir)
	t.Setenv("FILES_ENCRYPTED", "false")

	containerDir := filepath.Join(dir, "knowledge-nodes", "592_History")
	if err := os.MkdirAll(containerDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(containerDir, "history.mm"), []byte("<map/>"), 0o644); err != nil {
		t.Fatal(err)
	}

	svc := NewFilesService()
	files, err := svc.ListFilesInAbsoluteDir(containerDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Path != "knowledge-nodes/592_History/history.mm" || files[0].IsDir {
		t.Fatalf("unexpected files: %#v", files)
	}
}

func TestGetFilesFolderUsesFILES_DIR(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("FILES_DIR", dir)
	containerDir := filepath.Join(dir, "knowledge-nodes", "10_InteliJ")
	if err := os.MkdirAll(containerDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cs := &ContainerService{}
	found := cs.GetFilesFolder(t.Context(), schema.ContainerTypeKnowledgeNode, 10)
	if found == nil || *found != containerDir {
		t.Fatalf("expected %s, got %v", containerDir, found)
	}
}

func TestOwningContainerFromFilesRelPath(t *testing.T) {
	tests := []struct {
		path string
		typ  schema.ContainerType
		id   int
		ok   bool
	}{
		{"tasks/12_foo/note.md", schema.ContainerTypeTask, 12, true},
		{"knowledge-nodes/592_History/history.mm", schema.ContainerTypeKnowledgeNode, 592, true},
		{"/epics/5_name/a/b.txt", schema.ContainerTypeEpic, 5, true},
		{"readme.md", "", 0, false},
		{"tasks/note.md", "", 0, false},
		{"tasks/_12/note.md", "", 0, false},
		{"other/12_foo/note.md", "", 0, false},
	}
	for _, tt := range tests {
		gotType, gotID, ok := OwningContainerFromFilesRelPath(tt.path)
		if ok != tt.ok || gotType != tt.typ || gotID != tt.id {
			t.Fatalf("%q: got (%q, %d, %v), want (%q, %d, %v)",
				tt.path, gotType, gotID, ok, tt.typ, tt.id, tt.ok)
		}
	}
}
