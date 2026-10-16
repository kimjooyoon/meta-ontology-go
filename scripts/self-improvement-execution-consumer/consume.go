package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	continuation "github.com/kimjooyoon/meta-ontology-go/internal/meta/executioncontinuation"
)

func reserve(o options) error {
	var p continuation.Prepared
	if err := readJSON(o.input, &p); err != nil {
		return err
	}
	if err := currentInvocation(p, o); err != nil {
		return err
	}
	reservation, err := continuation.Reserve(os.DirFS("."), p)
	if err != nil {
		return err
	}
	return writeJSON(o.output, reservation)
}

func consume(o options) error {
	var p continuation.Prepared
	var r continuation.Reservation
	if err := readJSON(o.input, &p); err != nil {
		return err
	}
	if err := readJSON(o.reservation, &r); err != nil {
		return err
	}
	if err := currentInvocation(p, o); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(o.output), 0700); err != nil {
		return err
	}
	claim := filepath.Join(filepath.Dir(o.output), "execution-"+strings.ReplaceAll(r.Digest, ":", "-")+".claim")
	execution, err := continuation.Acquire(os.DirFS("."), p, r, claim)
	if err != nil {
		return err
	}
	report, err := execution.Run()
	if report.Schema != "" {
		if writeErr := writeJSON(o.output, report); writeErr != nil {
			return writeErr
		}
	}
	return err
}

func currentInvocation(p continuation.Prepared, o options) error {
	if p.Metadata.ConsumerRunID != o.run || p.Metadata.ConsumerAttempt != o.attempt || o.attempt != 1 {
		return fmt.Errorf("consumer invocation mismatch")
	}
	return nil
}
