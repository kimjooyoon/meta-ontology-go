package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	cfg := config{}
	flag.StringVar(&cfg.repository, "repository", "", "owner/repository")
	flag.StringVar(&cfg.predecessorSHA, "predecessor-sha", "", "exact predecessor SHA")
	flag.StringVar(&cfg.branch, "branch", "", "canonical branch")
	flag.StringVar(&cfg.workflow, "workflow", "", "canonical workflow name")
	flag.StringVar(&cfg.output, "output", "", "exclusive output path")
	flag.Parse()
	token := os.Getenv("GITHUB_TOKEN")
	baseURL := os.Getenv("GITHUB_API_URL")
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	if cfg.repository == "" || cfg.predecessorSHA == "" || cfg.branch == "" ||
		cfg.workflow == "" || cfg.output == "" || token == "" {
		fatal(fmt.Errorf("collector identity and GITHUB_TOKEN are required"))
	}
	result, err := collect(context.Background(), newGitHubClient(baseURL, token), cfg)
	if err != nil {
		fatal(err)
	}
	data, err := json.MarshalIndent(result.Input, "", "  ")
	if err != nil {
		fatal(err)
	}
	data = append(data, '\n')
	file, err := os.OpenFile(cfg.output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		fatal(err)
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		fatal(err)
	}
	if err := file.Close(); err != nil {
		fatal(err)
	}
	fmt.Printf("feedback-predecessor-input: sha=%s candidates=%d\n",
		cfg.predecessorSHA, len(result.Input.Candidates))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}
