package couplingmanifest

import (
	"errors"
	"testing"
)

func assertConstructionPartition(t *testing.T, output BuildOutput, err error, wantStatus ConstructionStatus, wantCode ConstructionCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected construction error, output=%#v", output)
	}
	var constructionErr *ConstructionError
	if !errors.As(err, &constructionErr) {
		t.Fatalf("error type = %T, want *ConstructionError: %v", err, err)
	}
	if constructionErr.Status != wantStatus || constructionErr.Code != wantCode {
		t.Fatalf("construction error = %#v, want status=%s code=%s", constructionErr, wantStatus, wantCode)
	}
	if output.Metadata.Status != wantStatus || output.Metadata.Reason != wantCode {
		t.Fatalf("metadata = %#v, want status=%s code=%s", output.Metadata, wantStatus, wantCode)
	}
	if output.Manifest.Complete || output.Manifest.Digest != "" || len(output.Manifest.Entries) != 0 {
		t.Fatalf("error output contains authoritative manifest data: %#v", output.Manifest)
	}
}
