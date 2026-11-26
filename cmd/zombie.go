package cmd

import (
	"fmt"
	"sort"
	"time"

	"bounty-analyzer/api"
)

func ListZombieReports(days int) error {
	if days <= 0 {
		return fmt.Errorf("days must be greater than zero")
	}

	cache := api.LoadCache()
	threshold := time.Now().AddDate(0, 0, -days)
	programNames := make([]string, 0, len(cache.Programs))
	for name := range cache.Programs {
		programNames = append(programNames, name)
	}
	sort.Strings(programNames)

	for _, name := range programNames {
		program := cache.Programs[name]
		reports := append([]api.Report(nil), program.Reports...)
		sort.SliceStable(reports, func(i, j int) bool {
			if reports[i].CreatedAt.Equal(reports[j].CreatedAt) {
				return reports[i].Title < reports[j].Title
			}
			return reports[i].CreatedAt.Before(reports[j].CreatedAt)
		})
		for _, report := range reports {
			if report.State == "new" && report.CreatedAt.Before(threshold) {
				fmt.Printf("Zombie Report: %s (%s) - Created: %s\n", report.Title, program.Name, report.CreatedAt)
			}
		}
	}

	return nil
}
