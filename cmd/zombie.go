package main

import (
	"fmt"
	"time"
	"bounty-analyzer/api"
)

func listZombieReports(days int) {
	cache := api.LoadCache()
	threshold := time.Now().AddDate(0, 0, -days)
	for _, program := range cache.Programs {
		for _, report := range program.Reports {
			if report.State == "new" && report.CreatedAt.Before(threshold) {
				fmt.Printf("Zombie Report: %s (%s) - Created: %s\n", report.Title, program.Name, report.CreatedAt)
			}
		}
	}
}
