package main

import (
	"context"
	"fmt"
)

func resolveParent(ctx context.Context, client *githubClient, repository,
	headSHA string) (string, error) {
	var value commit
	endpoint := fmt.Sprintf("/repos/%s/commits/%s", repository, headSHA)
	if err := client.getJSON(ctx, endpoint, &value); err != nil {
		return "", err
	}
	if value.SHA != headSHA || len(value.Parents) != 1 {
		return "", fmt.Errorf("ancestor does not have one canonical parent")
	}
	return value.Parents[0].SHA, nil
}
