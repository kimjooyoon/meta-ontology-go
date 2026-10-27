package prooftransport

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// The driver owns HTTP mechanics; its finite limits come from the Gooo plan.
type wire struct {
	client   *http.Client
	scope    R_TransportScope
	token    string
	requests int64
	bytes    int64
}

type response struct {
	status    int64
	data      []byte
	header    http.Header
	attempted bool
}

func digest(raw []byte) string { return fmt.Sprintf("sha256:%x", sha256.Sum256(raw)) }

func (w *wire) read(ctx context.Context, url string) (response, error) {
	return w.get(ctx, url, 0)
}

func (w *wire) get(ctx context.Context, target string, redirects int) (response, error) {
	res, err := w.send(ctx, http.MethodGet, target, nil)
	if err != nil || res.status < 300 || res.status > 399 {
		return res, err
	}
	if redirects >= 2 {
		return res, fmt.Errorf("redirect limit reached")
	}
	base, _ := url.Parse(target)
	next, err := base.Parse(res.header.Get("Location"))
	if err != nil || next.Scheme != "https" || next.User != nil {
		return res, fmt.Errorf("invalid read redirect")
	}
	return w.get(ctx, next.String(), redirects+1)
}

func (w *wire) send(ctx context.Context, method, url string, body io.Reader) (response, error) {
	if w.requests >= w.scope.F_max_requests {
		return response{}, fmt.Errorf("request budget exhausted")
	}
	remaining := w.scope.F_max_total_bytes - w.bytes
	if remaining <= 0 {
		return response{}, fmt.Errorf("response byte budget exhausted")
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return response{}, fmt.Errorf("invalid request")
	}
	if err := ctx.Err(); err != nil {
		return response{}, fmt.Errorf("request deadline")
	}
	req.Header.Set("Accept", "application/vnd.oci.image.index.v1+json, application/vnd.oci.image.manifest.v1+json, application/vnd.docker.distribution.manifest.v2+json")
	if method == http.MethodPut {
		req.Header.Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
	}
	base, _ := req.URL.Parse(w.scope.F_base_url)
	if w.token != "" && req.URL.Host == base.Host {
		req.Header.Set("Authorization", "Bearer "+w.token)
	}
	w.requests++
	res, err := w.client.Do(req)
	if err != nil {
		return response{attempted: true}, fmt.Errorf("HTTP request failed: %T", err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, min(w.scope.F_max_response_bytes, remaining)+1))
	w.bytes += int64(len(raw))
	out := response{status: int64(res.StatusCode), data: raw, header: res.Header, attempted: true}
	if int64(len(raw)) > w.scope.F_max_response_bytes || w.bytes > w.scope.F_max_total_bytes {
		return out, fmt.Errorf("response byte budget exhausted")
	}
	if err != nil {
		return out, fmt.Errorf("response read failed: %T", err)
	}
	return out, nil
}
