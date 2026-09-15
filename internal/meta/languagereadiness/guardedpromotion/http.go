package guardedpromotion

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Collector struct {
	APIURL string
	Token  string
	Client *http.Client
}

func NewCollector(apiURL, token string) *Collector {
	return &Collector{
		APIURL: strings.TrimRight(apiURL, "/"), Token: token,
		Client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (collector *Collector) getJSON(ctx context.Context, path string, target any) error {
	data, err := collector.get(ctx, path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

func (collector *Collector) get(ctx context.Context, path string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, collector.APIURL+path, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	request.Header.Set("Authorization", "Bearer "+collector.Token)
	response, err := collector.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("GET %s returned %s", path, response.Status)
	}
	return data, nil
}
