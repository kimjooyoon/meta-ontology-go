package analyzer

import (
	"errors"
	"testing"
)

func TestSourceBundleDigestRejectsDuplicateFilename(t *testing.T) {
	_, err := SourceBundleDigest([]SourceFile{
		{Filename: "same.go", Source: []byte("package p")},
		{Filename: "same.go", Source: []byte("package p")},
	})
	if !errors.Is(err, ErrSemanticAdapter) {
		t.Fatalf("duplicate source error = %v, want ErrSemanticAdapter", err)
	}
	var adapterErr AdapterError
	if !errors.As(err, &adapterErr) || adapterErr.Code != AdapterSourceConfig {
		t.Fatalf("duplicate source error = %v, want source-config", err)
	}
}
