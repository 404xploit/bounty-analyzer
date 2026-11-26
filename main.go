package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"bounty-analyzer/cmd"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: bounty-analyzer <analyze|zombie|dupes> [options]")
	}

	switch args[0] {
	case "analyze":
		flags := flag.NewFlagSet("analyze", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		programName := flags.String("program", "", "HackerOne program handle")
		if err := flags.Parse(args[1:]); err != nil {
			return fmt.Errorf("analyze: %w", err)
		}
		if flags.NArg() != 0 {
			return errors.New("analyze: unexpected positional arguments")
		}
		if *programName == "" {
			return errors.New("analyze: provide --program <handle>")
		}
		return cmd.AnalyzeProgram(*programName)

	case "zombie":
		flags := flag.NewFlagSet("zombie", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		days := flags.Int("days", 60, "Minimum age in days for a report to be considered stale")
		if err := flags.Parse(args[1:]); err != nil {
			return fmt.Errorf("zombie: %w", err)
		}
		if flags.NArg() != 0 {
			return errors.New("zombie: unexpected positional arguments")
		}
		if *days <= 0 {
			return errors.New("zombie: --days must be greater than zero")
		}
		return cmd.ListZombieReports(*days)

	case "dupes":
		flags := flag.NewFlagSet("dupes", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		keywords := flags.String("keywords", "", "Comma-separated report title keywords")
		if err := flags.Parse(args[1:]); err != nil {
			return fmt.Errorf("dupes: %w", err)
		}
		if flags.NArg() != 0 {
			return errors.New("dupes: unexpected positional arguments")
		}
		if *keywords == "" {
			return errors.New("dupes: provide --keywords <keyword[,keyword... ]>")
		}
		return cmd.DetectDupes(*keywords)

	default:
		return fmt.Errorf("unknown command %q (expected analyze, zombie, or dupes)", args[0])
	}
}
