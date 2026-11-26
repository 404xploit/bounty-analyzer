package cmd

import (
	"strings"
	"testing"

	"bounty-analyzer/api"
)

func TestFormatAnalysisSortsSeveritiesAndHandlesEmptyReports(t *testing.T) {
	program := api.Program{
		Name: "sample",
		Reports: []api.Report{
			{Severity: "medium", BountyAwarded: 50},
			{Severity: "critical", BountyAwarded: 100},
			{Severity: "high", BountyAwarded: 25},
		},
		DupePercentage: 10,
	}

	output := formatAnalysis(program)
	critical := strings.Index(output, "- critical:")
	high := strings.Index(output, "- high:")
	medium := strings.Index(output, "- medium:")
	if critical < 0 || high < 0 || medium < 0 || !(critical < high && high < medium) {
		t.Fatalf("severity output is not deterministic: %s", output)
	}

	emptyOutput := formatAnalysis(api.Program{Name: "empty"})
	if strings.Contains(strings.ToLower(emptyOutput), "nan") || strings.Contains(strings.ToLower(emptyOutput), "+inf") {
		t.Fatalf("empty report summary contains an invalid percentage: %s", emptyOutput)
	}
}
