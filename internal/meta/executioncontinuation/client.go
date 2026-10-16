package executioncontinuation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"time"
)

type Client struct {
	HTTP              *http.Client
	Repository, Token string
	Root              fs.FS
}

func (c Client) request(ctx context.Context, method, path string, body any) ([]byte, error) {
	if !repositoryName(c.Repository) {
		return nil, fmt.Errorf("invalid repository")
	}
	var data []byte
	if body != nil {
		var err error
		data, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, "https://api.github.com/repos/"+c.Repository+"/"+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2026-03-10")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	client := c.HTTP
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("workflow protocol request unavailable")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("workflow protocol HTTP %d", resp.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, MaxBytes+1))
	if err != nil || len(raw) > MaxBytes {
		return nil, fmt.Errorf("workflow protocol response unavailable or oversized")
	}
	return raw, nil
}

func decode(raw []byte, value any) error { return json.Unmarshal(raw, value) }

func (c Client) get(ctx context.Context, path string, value any) error {
	raw, err := c.request(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	return decode(raw, value)
}
