package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
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
	Handle         string   `json:"handle"`
	Name           string   `json:"name"`
	Reports        []Report `json:"reports"`
	AvgResponseTime float64  `json:"avg_response_time"`
	DupePercentage  float64  `json:"dupe_percentage"`
}

type Cache struct {
	Programs    map[string]Program `json:"programs"`
	LastUpdated time.Time          `json:"last_updated"`
}

const (
	hackerOneAPI = "https://hackerone.com"
	cacheFile    = "data/reports_cache.json"
)

func FetchProgramData(programName string) Program {
	var reports []Report
	var dupes int
	page := 1

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for {
		url := fmt.Sprintf("%s/%s/hacktivity.json?sort_type=latest_disclosable_activity_at&page=%d", hackerOneAPI, programName, page)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatalf("Erro ao criar requisição para página %d: %v", page, err)
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("Erro ao buscar dados da página %d: %v", page, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			log.Printf("Resposta inesperada na página %d: %d", page, resp.StatusCode)
			break
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Erro ao ler resposta da página %d: %v", page, err)
		}

		var result struct {
			Reports []struct {
				ID            string    `json:"id"`
				Title         string    `json:"title"`
				CreatedAt     time.Time `json:"disclosed_at"`
				Substate      string    `json:"substate"`
				SeverityRating string    `json:"severity_rating"`
				BountyAmount  float64   `json:"bounty_amount"`
			} `json:"reports"`
		}

		err = json.Unmarshal(body, &result)
		if err != nil {
			log.Fatalf("Erro ao parsear JSON da página %d: %v", page, err)
		}

		if len(result.Reports) == 0 {
			break
		}

		for _, r := range result.Reports {
			if r.Substate == "duplicate" {
				dupes++
			}
			reports = append(reports, Report{
				ID:            r.ID,
				Title:         r.Title,
				CreatedAt:     r.CreatedAt,
				State:         r.Substate,
				Severity:      r.SeverityRating,
				BountyAwarded: r.BountyAmount,
			})
		}

		page++
		time.Sleep(time.Duration(500+rand.Intn(500)) * time.Millisecond)
	}

	dupePercentage := 0.0
	if len(reports) > 0 {
		dupePercentage = float64(dupes) / float64(len(reports)) * 100
	}

	return Program{
		Handle:         programName,
		Name:           programName,
		Reports:        reports,
		AvgResponseTime: 0,
		DupePercentage: dupePercentage,
	}
}

func LoadCache() Cache {
	cache := Cache{Programs: make(map[string]Program)}
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		log.Printf("Aviso: cache não encontrado, iniciando vazio: %v", err)
		return cache
	}
	if err := json.Unmarshal(data, &cache); err != nil {
		log.Printf("Aviso: cache corrompido, iniciando vazio: %v", err)
		return cache
	}
	return cache
}

func SaveCache(cache Cache) {
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		log.Fatalf("Erro ao serializar cache: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(cacheFile), 0755); err != nil {
		log.Fatalf("Erro ao criar diretório de cache: %v", err)
	}
	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		log.Fatalf("Erro ao salvar cache: %v", err)
	}
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
