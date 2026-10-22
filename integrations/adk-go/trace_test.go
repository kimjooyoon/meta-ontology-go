package adkcheck

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestCompilerSpansExposeOnlyBoundedOutcomes(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() {
		otel.SetTracerProvider(previous)
		_ = provider.Shutdown(context.Background())
	})

	root := t.TempDir()
	fixture := filepath.Join(root, "fixture.sh")
	write := func(output string) {
		t.Helper()
		if err := os.WriteFile(fixture, []byte("#!/bin/sh\nprintf '"+output+"'\n"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	write(`{"status":"ok"}`)
	result, err := runJSONCommand(context.Background(), fixture, root, time.Second,
		"discover", "--query", "private-query", "private-source.gooo")
	if err != nil || result.ExitCode != 0 {
		t.Fatalf("successful compiler call: result=%#v err=%v", result, err)
	}
	write("invalid")
	_, err = runJSONCommand(context.Background(), fixture, root, time.Second,
		"check", "private-source.gooo")
	expectFailure(t, err, StageOutput, CodeInvalidJSON)

	spans := recorder.Ended()
	if len(spans) != 2 || spans[0].Name() != "gooo.discover" || spans[1].Name() != "gooo.check" {
		t.Fatalf("compiler spans = %v", spans)
	}
	if spans[0].Status().Code != codes.Unset || spans[1].Status().Code != codes.Error ||
		spans[1].Status().Description != "gooo-adk/output/invalid_json" {
		t.Fatalf("unexpected span statuses: %v, %v", spans[0].Status(), spans[1].Status())
	}
	for i, span := range spans {
		for _, attr := range span.Attributes() {
			if strings.Contains(string(attr.Key), "path") || strings.Contains(attr.Value.Emit(), "private") {
				t.Fatalf("span %d leaked caller input: %v", i, attr)
			}
		}
	}
}
