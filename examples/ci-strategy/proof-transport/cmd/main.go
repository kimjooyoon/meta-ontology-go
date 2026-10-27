package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/examples/ci-strategy/proof-transport"
)

func main() {
	if err := execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func execute() error {
	apply := flag.Bool("apply", false, "apply the single source-selected manifest restoration")
	materialPath := flag.String("material", "examples/ci-strategy/proof-transport/testdata/manifest.json", "original content-addressed manifest")
	flag.Parse()
	const operation = "transport-original-build-proof"
	scope := prooftransport.Plan(operation)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(scope.F_deadline_ms)*time.Millisecond)
	defer cancel()
	f, err := os.Open(*materialPath)
	if err != nil {
		return err
	}
	defer f.Close()
	material, err := io.ReadAll(io.LimitReader(f, scope.F_max_response_bytes+1))
	if err != nil {
		return err
	}
	token := os.Getenv("GOOO_REGISTRY_BEARER_TOKEN")
	if *apply && token == "" {
		return fmt.Errorf("apply requires GOOO_REGISTRY_BEARER_TOKEN from the execution environment")
	}
	if token == "" {
		token, err = anonymousToken(ctx, scope)
		if err != nil {
			return err
		}
	}
	report := prooftransport.Run(ctx, operation, material, token, *apply)
	if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
		return err
	}
	if report.Write.Attempted && (report.Write.Error != "" || report.Write.Status < 200 || report.Write.Status > 299) {
		return fmt.Errorf("manifest write did not return success; retain the separate readback result")
	}
	return nil
}

func anonymousToken(ctx context.Context, scope prooftransport.R_TransportScope) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, scope.F_token_url, nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("anonymous registry authentication failed")
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", fmt.Errorf("anonymous registry authentication returned %d", res.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(res.Body, scope.F_max_response_bytes+1))
	if err != nil || int64(len(raw)) > scope.F_max_response_bytes {
		return "", fmt.Errorf("anonymous token response unavailable or too large")
	}
	var value struct {
		Token string `json:"token"`
	}
	if json.Unmarshal(raw, &value) != nil || value.Token == "" {
		return "", fmt.Errorf("anonymous token response incomplete")
	}
	return value.Token, nil
}
