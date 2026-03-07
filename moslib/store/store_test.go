package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFSStore_VerifyReturnsNoErrors(t *testing.T) {
	fs := &FSStore{}
	errs, err := fs.Verify(".")
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("expected 0 integrity errors, got %d", len(errs))
	}
}

func TestFSStore_PackUnpackNoOp(t *testing.T) {
	fs := &FSStore{}
	if err := fs.Pack("."); err != nil {
		t.Fatalf("Pack: %v", err)
	}
	if err := fs.Unpack("."); err != nil {
		t.Fatalf("Unpack: %v", err)
	}
}

func TestAuditLogger_WritesLogEntries(t *testing.T) {
	root := t.TempDir()
	inner := &FSStore{}
	logger := NewAuditLogger(inner, root)

	testFile := filepath.Join(root, "test.txt")
	err := logger.WriteFile(testFile, []byte("hello"), 0o644)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	entries, err := ReadAuditLog(root)
	if err != nil {
		t.Fatalf("ReadAuditLog: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Operation != "write" {
		t.Errorf("expected operation=write, got %s", entries[0].Operation)
	}
	if entries[0].Path != testFile {
		t.Errorf("expected path=%s, got %s", testFile, entries[0].Path)
	}
}

func TestAuditLogger_RemoveLogsEntry(t *testing.T) {
	root := t.TempDir()
	inner := &FSStore{}
	logger := NewAuditLogger(inner, root)

	dir := filepath.Join(root, "subdir")
	os.MkdirAll(dir, 0o755)
	os.WriteFile(filepath.Join(dir, "f.txt"), []byte("x"), 0o644)

	err := logger.RemoveAll(dir)
	if err != nil {
		t.Fatalf("RemoveAll: %v", err)
	}

	entries, err := ReadAuditLog(root)
	if err != nil {
		t.Fatalf("ReadAuditLog: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Operation != "remove" {
		t.Errorf("expected operation=remove, got %s", entries[0].Operation)
	}
}

func TestReadAuditLog_Parses(t *testing.T) {
	root := t.TempDir()
	logDir := filepath.Join(root, ".mos", "store")
	os.MkdirAll(logDir, 0o755)
	os.WriteFile(filepath.Join(logDir, "audit.log"), []byte(
		"2026-03-01T00:00:00Z alice write /foo/bar abc123\n"+
			"2026-03-01T00:01:00Z bob remove /baz -\n",
	), 0o644)

	entries, err := ReadAuditLog(root)
	if err != nil {
		t.Fatalf("ReadAuditLog: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Actor != "alice" {
		t.Errorf("expected actor=alice, got %s", entries[0].Actor)
	}
	if entries[1].Operation != "remove" {
		t.Errorf("expected operation=remove, got %s", entries[1].Operation)
	}
}
