package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type httpFetcher struct {
	client   *http.Client
	maxBytes int64
}

func newHTTPFetcher() httpFetcher {
	client := &http.Client{
		Timeout: 20 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return errors.New("redirects are not allowed")
		},
	}
	return httpFetcher{client: client, maxBytes: 1 << 20}
}

func (fetcher httpFetcher) Fetch(ctx context.Context, sourceURL string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "gooo-source-authority-upstream/1")
	response, err := fetcher.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("source status %d", response.StatusCode)
	}
	content, err := io.ReadAll(io.LimitReader(response.Body, fetcher.maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(content)) > fetcher.maxBytes {
		return nil, errors.New("source exceeds byte limit")
	}
	return content, nil
}
