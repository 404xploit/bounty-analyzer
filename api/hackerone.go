package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Report struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	CreatedAt     time.Time `json:"disclosed_at"`
	State         string    `json:"substate"`
	Severity      string    `json:"severity_rating"`
	BountyAwarded float64   `json:"bounty_amount"`
}

type Program struct {
	Handle          string    `json:"handle"`
	Name            string    `json:"name"`
	Reports         []Report  `json:"reports"`
	AvgResponseTime float64   `json:"avg_response_time"`
	DupePercentage  float64   `json:"dupe_percentage"`
	UpdatedAt       time.Time `json:"updated_at,omitempty"`
}

type Cache struct {
	Programs    map[string]Program `json:"programs"`
	LastUpdated time.Time          `json:"last_updated"`
}

const (
	hackerOneAPI       = "https://hackerone.com"
	cacheFile          = "data/reports_cache.json"
	maxResponseBytes   = 8 << 20
	maxErrorBodyBytes  = 4 << 10
	requestTimeout     = 10 * time.Second
	paginationWaitTime = 500 * time.Millisecond
)

func FetchProgramData(programName string) (Program, error) {
	client := &http.Client{Timeout: requestTimeout}
	return FetchProgramDataWithClient(context.Background(), client, programName)
}

func FetchProgramDataWithClient(ctx context.Context, client *http.Client, programName string) (Program, error) {
	return fetchProgramData(ctx, client, hackerOneAPI, programName, waitForNextPage)
}

func fetchProgramData(ctx context.Context, client *http.Client, baseURL, programName string, wait func(context.Context) error) (Program, error) {
	handle := strings.TrimSpace(programName)
	if handle == "" {
		return Program{}, fmt.Errorf("program handle cannot be empty")
	}
	if handle == "." || handle == ".." || strings.ContainsAny(handle, "/\\?#") {
		return Program{}, fmt.Errorf("invalid program handle %q", programName)
	}
	if ctx == nil {
		return Program{}, fmt.Errorf("request context cannot be nil")
	}
	if client == nil {
		client = &http.Client{Timeout: requestTimeout}
	}

	endpoint := strings.TrimRight(baseURL, "/") + "/" + url.PathEscape(handle) + "/hacktivity.json"
	query := url.Values{}
	query.Set("sort_type", "latest_disclosable_activity_at")

	var reports []Report
	var duplicateCount int

	for page := 1; ; page++ {
		if page > 1 && wait != nil {
			if err := wait(ctx); err != nil {
				return Program{}, fmt.Errorf("wait before page %d: %w", page, err)
			}
		}

		query.Set("page", fmt.Sprintf("%d", page))
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+query.Encode(), nil)
		if err != nil {
			return Program{}, fmt.Errorf("create request for page %d: %w", page, err)
		}
		req.Header.Set("User-Agent", "bounty-analyzer/1.0")
		resp, err := client.Do(req)
		if err != nil {
			return Program{}, fmt.Errorf("request page %d: %w", page, err)
		}

		pageReports, readErr := decodeReportsPage(resp, page)
		closeErr := resp.Body.Close()
		if readErr != nil {
			return Program{}, readErr
		}
		if closeErr != nil {
			return Program{}, fmt.Errorf("close response body for page %d: %w", page, closeErr)
		}
		if len(pageReports) == 0 {
			break
		}

		for _, report := range pageReports {
			if strings.EqualFold(report.State, "duplicate") {
				duplicateCount++
			}
			reports = append(reports, report)
		}
	}

	dupePercentage := 0.0
	if len(reports) > 0 {
		dupePercentage = float64(duplicateCount) / float64(len(reports)) * 100
	}

	return Program{
		Handle:         handle,
		Name:           handle,
		Reports:        reports,
		DupePercentage: dupePercentage,
		UpdatedAt:      time.Now().UTC(),
	}, nil
}

func decodeReportsPage(resp *http.Response, page int) ([]Report, error) {
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxErrorBodyBytes))
		return nil, fmt.Errorf("HackerOne returned %s on page %d", resp.Status, page)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read response page %d: %w", page, err)
	}
	if len(body) > maxResponseBytes {
		return nil, fmt.Errorf("response page %d exceeds %d bytes", page, maxResponseBytes)
	}

	var result struct {
		Reports []struct {
			ID             string    `json:"id"`
			Title          string    `json:"title"`
			CreatedAt      time.Time `json:"disclosed_at"`
			Substate       string    `json:"substate"`
			SeverityRating string    `json:"severity_rating"`
			BountyAmount   float64   `json:"bounty_amount"`
		} `json:"reports"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response page %d: %w", page, err)
	}
	if result.Reports == nil {
		return nil, fmt.Errorf("parse response page %d: missing reports array", page)
	}

	reports := make([]Report, 0, len(result.Reports))
	for _, report := range result.Reports {
		reports = append(reports, Report{
			ID:            report.ID,
			Title:         report.Title,
			CreatedAt:     report.CreatedAt,
			State:         report.Substate,
			Severity:      report.SeverityRating,
			BountyAwarded: report.BountyAmount,
		})
	}
	return reports, nil
}

func waitForNextPage(ctx context.Context) error {
	timer := time.NewTimer(paginationWaitTime)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func LoadCache() Cache {
	return LoadCacheFrom(cacheFile)
}

func LoadCacheFrom(path string) Cache {
	cache := Cache{Programs: make(map[string]Program)}
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Aviso: cache não encontrado, iniciando vazio: %v", err)
		return cache
	}
	if err := json.Unmarshal(data, &cache); err != nil {
		log.Printf("Aviso: cache corrompido, iniciando vazio: %v", err)
		return Cache{Programs: make(map[string]Program)}
	}
	if cache.Programs == nil {
		cache.Programs = make(map[string]Program)
	}

	// Migrate older cache files that only stored one global timestamp.
	if !cache.LastUpdated.IsZero() {
		for name, program := range cache.Programs {
			if program.UpdatedAt.IsZero() {
				program.UpdatedAt = cache.LastUpdated
				cache.Programs[name] = program
			}
		}
	}
	return cache
}

func SaveCache(cache Cache) error {
	return SaveCacheTo(cacheFile, cache)
}

func SaveCacheTo(path string, cache Cache) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("cache path cannot be empty")
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("serialize cache: %w", err)
	}

	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0700); err != nil {
		return fmt.Errorf("create cache directory: %w", err)
	}

	temp, err := os.CreateTemp(directory, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temporary cache file: %w", err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)

	if err := temp.Chmod(0600); err != nil {
		_ = temp.Close()
		return fmt.Errorf("restrict cache file permissions: %w", err)
	}
	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return fmt.Errorf("write cache: %w", err)
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return fmt.Errorf("sync cache: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close cache: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace cache: %w", err)
	}
	return nil
}

func ProgramNeedsRefresh(program Program, exists bool, now time.Time, maxAge time.Duration) bool {
	if !exists || program.UpdatedAt.IsZero() || maxAge <= 0 {
		return true
	}
	return now.Sub(program.UpdatedAt) >= maxAge
}

func DetectDupes(keywords string, cache Cache) {
	kwList := strings.Split(keywords, ",")
	for _, program := range cache.Programs {
		for _, report := range program.Reports {
			for _, kw := range kwList {
				if strings.Contains(strings.ToLower(report.Title), strings.ToLower(kw)) {
					fmt.Printf("Report: %s (%s) - Keyword: %s\n", report.Title, program.Name, kw)
				}
			}
		}
	}
}
