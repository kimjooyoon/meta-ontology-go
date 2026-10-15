// Command candidate-delivery produces or consumes source-bound candidate bundles.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/candidatedelivery"
)

type options struct {
	mode, output, candidate, transport, event, subject, workflow, head, name string
	repository, request, resolution, verification                            string
	run                                                                      int64
	attempt                                                                  int
}

func main() {
	var o options
	flag.StringVar(&o.mode, "mode", "fetch", "produce, fetch, or eligibility")
	flag.StringVar(&o.output, "output", "", "caller-owned output directory or manifest file")
	flag.StringVar(&o.candidate, "candidate", "", "candidate report")
	flag.StringVar(&o.transport, "transport", "", "transport receipt")
	flag.StringVar(&o.event, "event", "", "producer event")
	flag.StringVar(&o.subject, "subject", "", "producer logical subject")
	flag.StringVar(&o.workflow, "workflow", "", "producer workflow commit")
	flag.StringVar(&o.head, "head", "", "producer run head")
	flag.StringVar(&o.name, "name", "", "producer artifact lookup label")
	flag.StringVar(&o.repository, "repository", "", "owner/repository")
	flag.Int64Var(&o.run, "run", 0, "producer run ID")
	flag.IntVar(&o.attempt, "attempt", 0, "producer run attempt")
	flag.StringVar(&o.request, "request", "", "authorization request")
	flag.StringVar(&o.resolution, "resolution", "", "authorization resolution")
	flag.StringVar(&o.verification, "verification", "", "authorization verification")
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := run(ctx, o); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, o options) error {
	if o.output == "" {
		return fmt.Errorf("-output is required")
	}
	switch o.mode {
	case "produce":
		candidate, err := os.ReadFile(o.candidate)
		if err != nil {
			return err
		}
		transport, err := os.ReadFile(o.transport)
		if err != nil {
			return err
		}
		manifest, err := candidatedelivery.Produce(candidatedelivery.Manifest{RunID: o.run, Attempt: o.attempt,
			Event: o.event, SubjectSHA: o.subject, WorkflowSHA: o.workflow, RunHeadSHA: o.head}, o.name,
			candidatedelivery.Bundle{Candidate: candidate, Transport: transport})
		if err != nil {
			return err
		}
		return writeJSON(o.output, manifest)
	case "fetch":
		return fetch(ctx, o)
	case "eligibility":
		request, _ := os.ReadFile(o.request)
		resolution, err := os.ReadFile(o.resolution)
		if err != nil {
			return err
		}
		verification, _ := os.ReadFile(o.verification)
		r := candidatedelivery.ObserveEligibility(request, resolution, verification)
		if err := writeJSON(o.output, r); err != nil {
			return err
		}
		if r.State == "REFUTED" {
			return fmt.Errorf("%s", r.Reason)
		}
		return nil
	default:
		return fmt.Errorf("unknown delivery mode %q", o.mode)
	}
}

func writeJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0600)
}
