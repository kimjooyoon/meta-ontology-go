package cache

import (
	"errors"
	"testing"
)

func TestNewKeyRejectsUnknownHostStage(t *testing.T) {
	_, err := NewKey(KeySpec{Namespace: "billing", HostStage: HostStage("future")})
	if !errors.Is(err, ErrInvalidHostStage) {
		t.Fatalf("unknown stage error = %v, want ErrInvalidHostStage", err)
	}
}
