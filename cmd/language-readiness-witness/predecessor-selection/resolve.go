package main

import (
	"context"
	"fmt"
)

func resolvePredecessor(ctx context.Context, client *githubClient, cfg config) (string, error) {
	if cfg.predecessor != "" {
		return cfg.predecessor, nil
	}
	var value commit
	endpoint := fmt.Sprintf("/repos/%s/commits/%s", cfg.repository, cfg.currentHead)
	if err := client.getJSON(ctx, endpoint, &value); err != nil {
		return "", err
	}
	if value.SHA != cfg.currentHead || len(value.Parents) != 1 {
		return "", fmt.Errorf("current head does not have one canonical predecessor")
	}
	return value.Parents[0].SHA, nil
}
