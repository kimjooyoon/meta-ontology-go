package main

import "flag"

type config struct {
	repository     string
	currentHeadSHA string
	sourceRunID    int64
	apiURL         string
	token          string
	out            string
	expectDecision string
}

func parseConfig() config {
	var value config
	flag.StringVar(&value.repository, "repository", "", "GitHub owner/repository")
	flag.StringVar(&value.currentHeadSHA, "current-head-sha", "", "current subject SHA")
	flag.Int64Var(&value.sourceRunID, "source-run-id", 0, "completed CI workflow run ID")
	flag.StringVar(&value.apiURL, "api-url", "https://api.github.com", "GitHub API URL")
	flag.StringVar(&value.token, "token", "", "GitHub Actions read token")
	flag.StringVar(&value.out, "out", "guarded-promotion.json", "report output path")
	flag.StringVar(&value.expectDecision, "expect-decision", "", "required decision")
	flag.Parse()
	return value
}
