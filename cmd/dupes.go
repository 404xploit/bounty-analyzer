package main

import (
	"bounty-analyzer/api"
)

func detectDupes(keywords string) {
	cache := api.LoadCache()
	api.DetectDupes(keywords, cache)
}