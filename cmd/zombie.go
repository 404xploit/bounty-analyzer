package cmd

import (
	"fmt"
	"time"

	"bounty-analyzer/api"
)

func ListZombieReports(days int) error {
	if days <= 0 {
		return fmt.Errorf("days must be greater than zero")
	}
	cache := api.LoadCache()
	threshold := time.Now().AddDate(0, 0, -days)
	for _, program := range cache.Programs {
		for _, report := range program.Reports {
			if report.State == "new" && report.CreatedAt.Before(threshold) {
				fmt.Printf("Zombie Report: %s (%s) - Created: %s\n", report.Title, program.Name, report.CreatedAt)
			}
		}
	}
	return nil
}
