package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollectFiles(t *testing.T) {
	dir := t.TempDir()

	// Create test structure
	file1 := filepath.Join(dir, "a.txt")
	os.WriteFile(file1, []byte("hello"), 0644)

	subdir := filepath.Join(dir, "sub")
	os.Mkdir(subdir, 0755)
	file2 := filepath.Join(subdir, "b.txt")
	os.WriteFile(file2, []byte("world"), 0644)

	// Non-regular file (symlink will be created later)
	file3 := filepath.Join(dir, "other.txt")
	os.WriteFile(file3, []byte("other"), 0644)

	var files []string
	visited := make(map[string]bool)
	collectFiles(dir, false, &files, visited, true)

	if len(files) != 3 {
		t.Fatalf("expected 3 files (a.txt, other.txt, sub/b.txt), got %d: %v", len(files), files)
	}
}

func TestCollectFiles_Recursive(t *testing.T) {
	dir := t.TempDir()

	file1 := filepath.Join(dir, "a.txt")
	os.WriteFile(file1, []byte("hello"), 0644)

	subdir := filepath.Join(dir, "sub")
	os.Mkdir(subdir, 0755)
	file2 := filepath.Join(subdir, "b.txt")
	os.WriteFile(file2, []byte("world"), 0644)

	var files []string
	visited := make(map[string]bool)
	collectFiles(dir, false, &files, visited, true)

	if len(files) != 2 {
		t.Fatalf("expected 2 files (a.txt, b.txt), got %d: %v", len(files), files)
	}

	hasB := false
	for _, f := range files {
		if f == file2 {
			hasB = true
			break
		}
	}
	if !hasB {
		t.Error("expected sub/b.txt to be collected recursively")
	}
}

func TestCollectFiles_Symlink(t *testing.T) {
	dir := t.TempDir()

	realdir := filepath.Join(dir, "real")
	os.Mkdir(realdir, 0755)
	realFile := filepath.Join(realdir, "data.txt")
	os.WriteFile(realFile, []byte("data"), 0644)

	link := filepath.Join(dir, "link")
	err := os.Symlink(realdir, link)
	if err != nil {
		t.Skip("symlinks not supported:", err)
	}

	var files []string
	visited := make(map[string]bool)

	// Without following symlinks, the symlink should be skipped
	collectFiles(link, false, &files, visited, true)
	// collectFiles calls os.Lstat which will see a symlink dir,
	// then since !followSymlinks, it will return (skip)
	if len(files) != 0 {
		t.Errorf("expected 0 files when not following symlinks, got %d", len(files))
	}

	// With following symlinks
	files = nil
	visited = make(map[string]bool)
	collectFiles(link, true, &files, visited, true)
	if len(files) != 1 {
		t.Errorf("expected 1 file when following symlinks, got %d: %v", len(files), files)
	}
}

func TestCollectFiles_NonExistent(t *testing.T) {
	var files []string
	visited := make(map[string]bool)
	// Should not panic or error out (noWarn=true)
	collectFiles("/nonexistent/path", false, &files, visited, true)
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestCollectFiles_RegularFileOnly(t *testing.T) {
	dir := t.TempDir()

	// Create a regular file and a directory (no regular files inside)
	filepath.Join(dir, "empty_sub")

	os.WriteFile(filepath.Join(dir, "only.txt"), []byte("content"), 0644)

	var files []string
	visited := make(map[string]bool)
	collectFiles(dir, false, &files, visited, true)

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d: %v", len(files), files)
	}
	if filepath.Base(files[0]) != "only.txt" {
		t.Errorf("expected only.txt, got %s", filepath.Base(files[0]))
	}
}
