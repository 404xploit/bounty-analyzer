package cmd

import (
	"fmt"
	"time"

	"bounty-analyzer/api"
)

func AnalyzeProgram(programName string) error {
	cache := api.LoadCache()
	program, exists := cache.Programs[programName]
	if !exists || time.Since(cache.LastUpdated).Hours() > 24 {
		var err error
		program, err = api.FetchProgramData(programName)
		if err != nil {
			return fmt.Errorf("fetch program %q: %w", programName, err)
		}
		cache.Programs[programName] = program
		cache.LastUpdated = time.Now()
		api.SaveCache(cache)
	}

	var totalBounty float64
	severityCount := make(map[string]int)
	for _, report := range program.Reports {
		totalBounty += report.BountyAwarded
		if report.Severity != "" {
			severityCount[report.Severity]++
		}
	}

	fmt.Printf("=== Análise do Programa: %s ===\n", program.Name)
	fmt.Printf("Total de Reports: %d\n", len(program.Reports))
	fmt.Printf("Porcentagem de Duplicatas: %.2f%%\n", program.DupePercentage)
	fmt.Printf("Total Pago em Bounties: $%.2f\n", totalBounty)
	fmt.Printf("Distribuição de Severidade:\n")
	for severity, count := range severityCount {
		fmt.Printf("  - %s: %d (%.2f%%)\n", severity, count, float64(count)/float64(len(program.Reports))*100)
	}
	return nil
}
