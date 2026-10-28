package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/examples/ci-strategy/proof-transport"
	"github.com/kimjooyoon/meta-ontology-go/examples/ci-strategy/transport-origins"
)

func main() {
	if err := execute(os.Args[1:], os.Stdout, os.Stderr); err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func execute(args []string, output, diagnostics io.Writer) error {
	flags := flag.NewFlagSet("proof-transport", flag.ContinueOnError)
	flags.SetOutput(diagnostics)
	apply := flags.Bool("apply", false, "apply the single source-selected manifest restoration")
	materialPath := flags.String("material", "examples/ci-strategy/proof-transport/testdata/manifest.json",
		"original content-addressed manifest")
	prior := flags.String("reassess", "", "query a saved transport report without contacting the registry")
	current := flags.String("read-report", "", "saved read-only report to pair with -reassess")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected argument %q", flags.Arg(0))
	}
	priorSet, currentSet := false, false
	flags.Visit(func(option *flag.Flag) {
		priorSet = priorSet || option.Name == "reassess"
		currentSet = currentSet || option.Name == "read-report"
	})
	if priorSet {
		if *prior == "" || currentSet && *current == "" {
			return fmt.Errorf("saved-report query requires nonempty report paths")
		}
		if *apply {
			return fmt.Errorf("-reassess and -apply select different operations")
		}
		return reassessFiles(*prior, *current, output)
	}
	if currentSet {
		return fmt.Errorf("-read-report requires -reassess")
	}
	return executeTransport(*apply, *materialPath, output)
}

func executeTransport(apply bool, materialPath string, output io.Writer) error {
	const operation = "transport-original-build-proof"
	scope := prooftransport.Plan(operation)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(scope.F_deadline_ms)*time.Millisecond)
	defer cancel()
	material, err := readBoundedFile(materialPath, scope.F_max_response_bytes)
	if err != nil {
		return err
	}
	token := os.Getenv("GOOO_REGISTRY_BEARER_TOKEN")
	if apply && token == "" {
		return fmt.Errorf("apply requires GOOO_REGISTRY_BEARER_TOKEN from the execution environment")
	}
	if token == "" {
		token, err = anonymousToken(ctx, scope)
		if err != nil {
			return err
		}
	}
	report := prooftransport.Run(ctx, operation, material, token, apply)
	if err := json.NewEncoder(output).Encode(report); err != nil {
		return err
	}
	if report.Outcome.F_command_failed {
		return fmt.Errorf("operation outcome %s; retain the separate response and readback", report.Outcome.F_state)
	}
	return nil
}

func reassessFiles(priorPath, currentPath string, output io.Writer) error {
	limit := transportorigins.ReportByteLimit()
	prior, err := readBoundedFile(priorPath, limit)
	if err != nil {
		return err
	}
	var current []byte
	if currentPath != "" {
		current, err = readBoundedFile(currentPath, limit)
		if err != nil {
			return err
		}
	}
	result, err := transportorigins.ReassessSavedReports(prior, current)
	if err != nil {
		return err
	}
	return json.NewEncoder(output).Encode(result)
}

func readBoundedFile(path string, limit int64) ([]byte, error) {
	if limit < 1 || limit == math.MaxInt64 {
		return nil, fmt.Errorf("file byte limit cannot be represented")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(io.LimitReader(f, limit+1))
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
