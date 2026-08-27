package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	root, repository, currentHead, predecessorSHA string
	output, check, apiURL, token                  string
	observationCapture, observationReplay         string
}

func main() {
	cfg := config{apiURL: os.Getenv("GITHUB_API_URL"), token: os.Getenv("GITHUB_TOKEN")}
	if cfg.apiURL == "" {
		cfg.apiURL = "https://api.github.com"
	}
	flag.StringVar(&cfg.root, "root", "", "repository root")
	flag.StringVar(&cfg.repository, "repository", "", "owner/name repository")
	flag.StringVar(&cfg.currentHead, "current-head", "", "current exact commit sha")
	flag.StringVar(&cfg.predecessorSHA, "predecessor-sha", "", "merged evidence commit sha")
	flag.StringVar(&cfg.output, "output", "", "new receipt outside the repository")
	flag.StringVar(&cfg.check, "check", "", "existing receipt outside the repository")
	flag.StringVar(&cfg.observationCapture, "observation-capture", "", "raw GitHub observation cache outside the repository")
	flag.StringVar(&cfg.observationReplay, "observation-replay", "", "raw GitHub observation cache to replay")
	flag.Parse()
	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
