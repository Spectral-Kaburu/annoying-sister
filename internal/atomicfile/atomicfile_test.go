package atomicfile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Spectral-Kaburu/annoying-sister/internal/atomicfile"
)

func TestWrite_CreatesFileWithContent(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "test.txt")
	content := []byte("hello world")

	if err := atomicfile.Write(target, content, 0o644); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	read, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(read) != string(content) {
		t.Errorf("got %q, want %q", string(read), string(content))
	}
}

func TestWrite_SetsPermissions(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "exec.sh")
	content := []byte("#!/bin/sh\necho hi\n")

	if err := atomicfile.Write(target, content, 0o755); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	if info.Mode().Perm() != 0o755 {
		t.Errorf("got perm %o, want 0755", info.Mode().Perm())
	}
}

func TestWrite_OverwritesAtomically(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.json")

	if err := atomicfile.Write(target, []byte("initial"), 0o644); err != nil {
		t.Fatalf("initial Write failed: %v", err)
	}

	if err := atomicfile.Write(target, []byte("overwritten"), 0o644); err != nil {
		t.Fatalf("second Write failed: %v", err)
	}

	read, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(read) != "overwritten" {
		t.Errorf("got %q, want 'overwritten'", string(read))
	}
}

func TestWrite_ErrorOnInvalidDirectory(t *testing.T) {
	target := filepath.Join(t.TempDir(), "nonexistent_dir", "file.txt")
	err := atomicfile.Write(target, []byte("fail"), 0o644)
	if err == nil {
		t.Fatal("expected error when writing to non-existent directory, got nil")
	}
}
