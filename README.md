# Bounty Analyzer

Bounty Analyzer is a small Go CLI for reviewing public HackerOne Hacktivity data from a local cache. It is intended for authorized bug bounty research and reporting workflows.

## Requirements

- Go 1.22 or newer
- Network access to `hackerone.com` when refreshing a program

## Build and test

```bash
go test ./...
go build -o bounty-analyzer .
```

## Commands

Analyze a program and refresh its cached public report data when the entry is older than 24 hours:

```bash
./bounty-analyzer analyze --program khealth
```

List reports still marked `new` whose disclosure timestamp is older than the selected number of days (default: 60):

```bash
./bounty-analyzer zombie --days 60
```

Find report titles containing one or more comma-separated keywords in the local cache:

```bash
./bounty-analyzer dupes --keywords "graphql,idor,ssrf"
```

## Cache behavior

Cached data is stored at `data/reports_cache.json`. Each program has its own refresh timestamp, so refreshing one program does not make other cached programs appear fresh. Cache updates use a temporary file and atomic rename; the cache file is created with owner-only permissions.

The `analyze` command retrieves public Hacktivity data. It does not authenticate to HackerOne or access private program reports.

## Responsible use

Only use this tool for programs and information you are authorized to review. Respect HackerOne's terms, rate limits, and applicable disclosure requirements.
