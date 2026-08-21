package cache

import (
	"errors"
	"testing"
)

func TestGetProjectionIfFreshRejectsReplayedReceipt(t *testing.T) {
	cache, key, identity, evidence, receipt := projectionHitFixture(t)
	sealed, err := receipt.Seal()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cache.AppendReceipt(sealed); err != nil {
		t.Fatal(err)
	}
	if _, err := cache.AppendReceipt(sealed); !errors.Is(err, ErrReceiptReplay) {
		t.Fatalf("replayed receipt append = %v, want ErrReceiptReplay", err)
	}
	if _, _, err := cache.GetProjectionIfFresh(key, identity, evidence, sealed); err != nil {
		t.Fatalf("single recorded receipt should remain usable: %v", err)
	}
}
func TestGetProjectionIfFreshRejectsUnknownFreshness(t *testing.T) {
	cache, key, identity, evidence, receipt := projectionHitFixture(t)
	sealed, err := receipt.Seal()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cache.AppendReceipt(sealed); err != nil {
		t.Fatal(err)
	}
	for name, invoke := range map[string]func() error{
		"unknown identity": func() error {
			unknown := identity
			unknown.SourceDigest = ""
			_, _, err := cache.GetProjectionIfFresh(key, unknown, evidence, sealed)
			return err
		},
		"unknown current": func() error {
			unknown := evidence
			unknown.SourceDigest = ""
			_, _, err := cache.GetProjectionIfFresh(key, identity, unknown, sealed)
			return err
		},
		"unknown receipt": func() error {
			unknown := sealed
			unknown.Evidence.PolicyDigest = ""
			_, _, err := cache.GetProjectionIfFresh(key, identity, evidence, unknown)
			return err
		},
	} {
		t.Run(name, func(t *testing.T) {
			err := invoke()
			if !errors.Is(err, ErrUnknownFreshness) && !errors.Is(err, ErrInvalidReceipt) {
				t.Fatalf("unknown freshness = %v", err)
			}
			data, metadata, err := cache.Get(key)
			if err != nil || string(data) != "projection" || metadata.Size != int64(len(data)) {
				t.Fatalf("unknown freshness changed projection: %q %+v %v", data, metadata, err)
			}
		})
	}
}
