package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateDir(t *testing.T) {
	t.Parallel()

	t.Run("creates missing dir", func(t *testing.T) {
		t.Parallel()

		dir := filepath.Join(t.TempDir(), "new-dir")
		if err := CreateDir(dir); err != nil {
			t.Fatalf("CreateDir failed: %v", err)
		}

		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("stat failed: %v", err)
		}

		if !info.IsDir() {
			t.Fatalf("expected %q to be a directory", dir)
		}
	})

	t.Run("clears existing dir", func(t *testing.T) {
		t.Parallel()

		dir := filepath.Join(t.TempDir(), "existing-dir")
		if err := os.MkdirAll(filepath.Join(dir, "nested"), 0o700); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}

		if err := os.WriteFile(filepath.Join(dir, "nested", "file.txt"), []byte("x"), 0o600); err != nil {
			t.Fatalf("write failed: %v", err)
		}

		if err := CreateDir(dir); err != nil {
			t.Fatalf("CreateDir failed: %v", err)
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("readdir failed: %v", err)
		}

		if len(entries) != 0 {
			t.Fatalf("expected directory to be empty, got %d entries", len(entries))
		}
	})

	t.Run("replaces file with dir", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "file-path")
		if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
			t.Fatalf("write failed: %v", err)
		}

		if err := CreateDir(path); err != nil {
			t.Fatalf("CreateDir failed: %v", err)
		}

		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat failed: %v", err)
		}

		if !info.IsDir() {
			t.Fatalf("expected %q to be a directory", path)
		}
	})
}

func TestClearDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o600); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "nested"), 0o700); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "nested", "b.txt"), []byte("b"), 0o600); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	if err := ClearDir(dir); err != nil {
		t.Fatalf("ClearDir failed: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir failed: %v", err)
	}

	if len(entries) != 0 {
		t.Fatalf("expected empty directory, got %d entries", len(entries))
	}
}

func TestFormatDirName(t *testing.T) {
	t.Parallel()

	got := FormatDirName("/a/b/c/")
	if got != "a@b@c" {
		t.Fatalf("unexpected result: got %q", got)
	}
}

func TestFileStat(t *testing.T) {
	t.Parallel()

	t.Run("existing file", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "file.txt")
		if err := os.WriteFile(path, []byte("abc"), 0o600); err != nil {
			t.Fatalf("write failed: %v", err)
		}

		stat, exists, err := FileStat(path)
		if err != nil {
			t.Fatalf("FileStat failed: %v", err)
		}

		if !exists || stat == nil {
			t.Fatalf("expected existing file stat")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "missing.txt")

		stat, exists, err := FileStat(path)
		if err != nil {
			t.Fatalf("FileStat failed: %v", err)
		}

		if exists || stat != nil {
			t.Fatalf("expected missing file")
		}
	})
}
