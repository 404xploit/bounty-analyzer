package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: bounty-analyzer <command> [options]")
		fmt.Println("Commands: analyze, zombie, dupes")
		fmt.Println("Example: bounty-analyzer analyze --program khealth")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "analyze":
		analyzeCmd := flag.NewFlagSet("analyze", flag.ExitOnError)
		programName := analyzeCmd.String("program", "", "Program to analyze (e.g., khealth)")
		analyzeCmd.Parse(os.Args[2:])
		if *programName == "" {
			log.Fatal("Please provide a program name with --program")
		}
		analyzeProgram(*programName)
	case "zombie":
		zombieCmd := flag.NewFlagSet("zombie", flag.ExitOnError)
		days := zombieCmd.Int("days", 60, "Days to consider a report as zombie")
		zombieCmd.Parse(os.Args[2:])
		listZombieReports(*days)
	case "dupes":
		dupesCmd := flag.NewFlagSet("dupes", flag.ExitOnError)
		keywords := dupesCmd.String("keywords", "", "Comma-separated keywords (e.g., graphql,idor)")
		dupesCmd.Parse(os.Args[2:])
		if *keywords == "" {
			log.Fatal("Please provide keywords with --keywords")
		}
		detectDupes(*keywords)
	default:
		fmt.Println("Unknown command:", os.Args[1])
		os.Exit(1)
	}
}