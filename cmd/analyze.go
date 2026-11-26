package cmd

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"

	"bounty-analyzer/api"
)

const cacheMaxAge = 24 * time.Hour

func AnalyzeProgram(programName string) error {
	programName = strings.TrimSpace(programName)
	if programName == "" {
		return fmt.Errorf("program handle cannot be empty")
	}

	cache := api.LoadCache()
	program, exists := cache.Programs[programName]
	if api.ProgramNeedsRefresh(program, exists, time.Now(), cacheMaxAge) {
		var err error
		program, err = api.FetchProgramData(programName)
		if err != nil {
			return fmt.Errorf("fetch program %q: %w", programName, err)
		}
		cache.Programs[programName] = program
		cache.LastUpdated = program.UpdatedAt
		if err := api.SaveCache(cache); err != nil {
			return fmt.Errorf("save report cache: %w", err)
		}
	}

	fmt.Print(formatAnalysis(program))
	return nil
}

func formatAnalysis(program api.Program) string {
	var output bytes.Buffer
	var totalBounty float64
	severityCount := make(map[string]int)
	for _, report := range program.Reports {
		totalBounty += report.BountyAwarded
		if severity := report.Severity; severity != "" {
			severityCount[severity]++
		}
	}

	fmt.Fprintf(&output, "=== Análise do Programa: %s ===\n", program.Name)
	fmt.Fprintf(&output, "Total de Reports: %d\n", len(program.Reports))
	fmt.Fprintf(&output, "Porcentagem de Duplicatas: %.2f%%\n", program.DupePercentage)
	fmt.Fprintf(&output, "Total Pago em Bounties: $%.2f\n", totalBounty)
	fmt.Fprintln(&output, "Distribuição de Severidade:")

	severities := make([]string, 0, len(severityCount))
	for severity := range severityCount {
		severities = append(severities, severity)
	}
	sort.Strings(severities)

	for _, severity := range severities {
		percentage := 0.0
		if len(program.Reports) > 0 {
			percentage = float64(severityCount[severity]) / float64(len(program.Reports)) * 100
		}
		fmt.Fprintf(&output, "  - %s: %d (%.2f%%)\n", severity, severityCount[severity], percentage)
	}

	return output.String()
}
