package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchProgramDataPaginatesAndCountsDuplicates(t *testing.T) {
	var requestedPages []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sample-program/hacktivity.json" {
			t.Errorf("unexpected request path: %s", r.URL.Path)
		}
		if got := r.Header.Get("User-Agent"); got != "bounty-analyzer/1.0" {
			t.Errorf("unexpected user agent: %q", got)
		}
		page := r.URL.Query().Get("page")
		requestedPages = append(requestedPages, page)
		switch page {
		case "1":
			fmt.Fprint(w, `{"reports":[{"id":"1","title":"IDOR","disclosed_at":"2026-01-02T03:04:05Z","substate":"duplicate","severity_rating":"high","bounty_amount":125.5}]}`)
		case "2":
			fmt.Fprint(w, `{"reports":[]}`)
		default:
			t.Errorf("unexpected page requested: %s", page)
			http.Error(w, "unexpected page", http.StatusBadRequest)
		}
	}))
	defer server.Close()

	program, err := fetchProgramData(context.Background(), server.Client(), server.URL, " sample-program ", nil)
	if err != nil {
		t.Fatalf("fetchProgramData returned error: %v", err)
	}
	if program.Handle != "sample-program" || program.Name != "sample-program" {
		t.Fatalf("unexpected program identity: %#v", program)
	}
	if len(program.Reports) != 1 || program.Reports[0].ID != "1" {
		t.Fatalf("unexpected reports: %#v", program.Reports)
	}
	if program.DupePercentage != 100 {
		t.Fatalf("expected 100%% duplicates, got %.2f", program.DupePercentage)
	}
	if strings.Join(requestedPages, ",") != "1,2" {
		t.Fatalf("expected pages 1 and 2, got %v", requestedPages)
	}
}

func TestFetchProgramDataRejectsInvalidHandles(t *testing.T) {
	for _, handle := range []string{"", "  ", "../other", "a/b", "a?b", "a#b"} {
		t.Run(fmt.Sprintf("%q", handle), func(t *testing.T) {
			_, err := fetchProgramData(context.Background(), http.DefaultClient, "https://example.test", handle, nil)
			if err == nil {
				t.Fatalf("expected invalid handle %q to fail", handle)
			}
		})
	}
}

func TestFetchProgramDataReturnsHTTPAndJSONErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		want       string
	}{
		{name: "http status", statusCode: http.StatusTooManyRequests, body: "slow down", want: "429 Too Many Requests"},
		{name: "invalid json", statusCode: http.StatusOK, body: "not-json", want: "parse response page 1"},
		{name: "missing reports field", statusCode: http.StatusOK, body: `{"data":[]}`, want: "missing reports array"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(test.statusCode)
				fmt.Fprint(w, test.body)
			}))
			defer server.Close()

			_, err := fetchProgramData(context.Background(), server.Client(), server.URL, "sample", nil)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected error containing %q, got %v", test.want, err)
			}
		})
	}
}

func TestFetchProgramDataRejectsOversizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"reports":[]}`))
		_, _ = w.Write([]byte(strings.Repeat("x", maxResponseBytes+1)))
	}))
	defer server.Close()

	_, err := fetchProgramData(context.Background(), server.Client(), server.URL, "sample", nil)
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected oversized response error, got %v", err)
	}
}

func TestFetchProgramDataHonorsCanceledContextBetweenPages(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "1" {
			fmt.Fprint(w, `{"reports":[{"id":"1","title":"first"}]}`)
			cancel()
			return
		}
		fmt.Fprint(w, `{"reports":[]}`)
	}))
	defer server.Close()

	_, err := fetchProgramData(ctx, server.Client(), server.URL, "sample", func(ctx context.Context) error {
		return ctx.Err()
	})
	if err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("expected canceled context to fail before page 2, got %v", err)
	}
}
