package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsFileOrDirExists(t *testing.T) {
	tmp := t.TempDir()

	// Existing directory.
	if !IsFileOrDirExists(tmp) {
		t.Errorf("expected existing dir %q to return true", tmp)
	}

	// Existing file.
	f := filepath.Join(tmp, "test.txt")
	if err := os.WriteFile(f, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if !IsFileOrDirExists(f) {
		t.Errorf("expected existing file %q to return true", f)
	}

	// Non-existent path.
	nonExistent := filepath.Join(tmp, "does-not-exist")
	if IsFileOrDirExists(nonExistent) {
		t.Errorf("expected non-existent path %q to return false", nonExistent)
	}
}
