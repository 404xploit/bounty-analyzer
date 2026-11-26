package cmd

import "bounty-analyzer/api"

func DetectDupes(keywords string) error {
	api.DetectDupes(keywords, api.LoadCache())
	return nil
}
