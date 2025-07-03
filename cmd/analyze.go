package main

import (
	"fmt"
	"bounty-analyzer/api"
)

func analyzeProgram(programName string) {
	cache := api.LoadCache()
	program, exists := cache.Programs[programName]
	if !exists || time.Since(cache.LastUpdated).Hours() > 24 {
		program = api.FetchProgramData(programName)
		cache.Programs[programName] = program
		cache.LastUpdated = time.Now()
		api.SaveCache(cache)
	}

	var totalBounty float64
	severityCount := make(map[string]int)
	for _, r := range program.Reports {
		totalBounty += r.BountyAwarded
		if r.Severity != "" {
			severityCount[r.Severity]++
		}
	}

	fmt.Printf("=== Análise do Programa: %s ===\n", program.Name)
	fmt.Printf("Total de Reports: %d\n", len(program.Reports))
	fmt.Printf("Porcentagem de Duplicatas: %.2f%%\n", program.DupePercentage)
	fmt.Printf("Total Pago em Bounties: $%.2f\n", totalBounty)
	fmt.Printf("Distribuição de Severidade:\n")
	for sev, count := range severityCount {
		fmt.Printf("  - %s: %d (%.2f%%)\n", sev, count, float64(count)/float64(len(program.Reports))*100)
	}
}