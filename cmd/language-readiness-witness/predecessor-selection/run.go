package main

import (
	"context"
	"fmt"
	"os"
)

func run(cfg config) error {
	if cfg.root == "" || cfg.repository == "" || cfg.currentHead == "" ||
		cfg.branch == "" || cfg.workflow == "" || cfg.baseline == "" ||
		cfg.reference == "" || cfg.bindingBaseline == "" || cfg.receipt == "" {
		return fmt.Errorf("selection identity and outputs are required")
	}
	for _, path := range []string{cfg.baseline, cfg.reference, cfg.bindingBaseline, cfg.receipt} {
		if err := requireOutside(cfg.root, path); err != nil {
			return err
		}
	}
	token, baseURL := os.Getenv("GITHUB_TOKEN"), os.Getenv("GITHUB_API_URL")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN is required")
	}
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	ctx := context.Background()
	client := newGitHubClient(baseURL, token)
	predecessor, err := resolvePredecessor(ctx, client, cfg)
	if err != nil {
		return err
	}
	result, resolution, err := resolveAncestry(ctx, client, cfg, predecessor)
	if err != nil {
		return err
	}
	return writeResolutionResult(cfg, result, resolution)
}
