package api

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveCacheToWritesAtomicallyWithPrivatePermissions(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "nested", "cache.json")
	cache := Cache{
		Programs: map[string]Program{
			"sample": {Name: "sample", UpdatedAt: time.Date(2026, 11, 25, 0, 0, 0, 0, time.UTC)},
		},
	}
	if err := SaveCacheTo(path, cache); err != nil {
		t.Fatalf("SaveCacheTo returned error: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("expected cache permissions 0600, got %04o", got)
	}
	loaded := LoadCacheFrom(path)
	if _, exists := loaded.Programs["sample"]; !exists {
		t.Fatalf("saved cache did not round-trip: %#v", loaded)
	}

	matches, err := filepath.Glob(filepath.Join(filepath.Dir(path), ".cache.json.tmp-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("temporary cache files were not cleaned up: %v", matches)
	}
}

func TestSaveCacheToRejectsEmptyPath(t *testing.T) {
	if err := SaveCacheTo("  ", Cache{}); err == nil {
		t.Fatal("expected empty path to fail")
	}
}
