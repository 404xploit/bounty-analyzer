package api

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestProgramNeedsRefreshUsesPerProgramTimestamp(t *testing.T) {
	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	maxAge := 24 * time.Hour

	fresh := Program{UpdatedAt: now.Add(-time.Hour)}
	stale := Program{UpdatedAt: now.Add(-48 * time.Hour)}
	if ProgramNeedsRefresh(fresh, true, now, maxAge) {
		t.Fatal("fresh program should not be refreshed")
	}
	if !ProgramNeedsRefresh(stale, true, now, maxAge) {
		t.Fatal("stale program should be refreshed")
	}
	if !ProgramNeedsRefresh(Program{}, true, now, maxAge) {
		t.Fatal("program without timestamp should be refreshed")
	}
	if !ProgramNeedsRefresh(fresh, false, now, maxAge) {
		t.Fatal("missing program should be refreshed")
	}
}

func TestLoadCacheBackfillsLegacyProgramTimestamps(t *testing.T) {
	updatedAt := time.Date(2026, 10, 1, 8, 30, 0, 0, time.UTC)
	content := fmt.Sprintf(`{"last_updated":%q,"programs":{"one":{"name":"one"},"two":{"name":"two","updated_at":"2026-09-01T00:00:00Z"}}}`, updatedAt.Format(time.RFC3339))
	path := t.TempDir() + "/cache.json"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	cache := LoadCacheFrom(path)
	if !cache.Programs["one"].UpdatedAt.Equal(updatedAt) {
		t.Fatalf("legacy timestamp was not backfilled: %s", cache.Programs["one"].UpdatedAt)
	}
	if got := cache.Programs["two"].UpdatedAt.Format(time.RFC3339); got != "2026-09-01T00:00:00Z" {
		t.Fatalf("existing per-program timestamp was overwritten: %s", got)
	}
}
