package main

import (
	"strings"
	"testing"
)

func TestCICacheC1KeyMutationFailsClosed(t *testing.T) {
	cache := validCache()
	cache.Key = "mutated"
	if err := validateCache(cache, evidenceInput{HeadSHA: strings.Repeat("a", 40)}); err == nil {
		t.Fatal("cache key mutation was accepted")
	}
}

func TestCICacheC2ContentSizeMismatchFailsClosed(t *testing.T) {
	cache := validCache()
	cache.HitContentSize++
	if err := validateCache(cache, evidenceInput{HeadSHA: strings.Repeat("a", 40)}); err == nil {
		t.Fatal("cache content-size mismatch was accepted")
	}
}

func TestCICacheC3UnknownDependencyFailsClosed(t *testing.T) {
	cache := validCache()
	cache.DirectDependencies = nil
	if err := validateCache(cache, evidenceInput{HeadSHA: strings.Repeat("a", 40)}); err == nil {
		t.Fatal("unknown dependency evidence was accepted")
	}
}

func TestCICacheC4ReplayPredecessorFailsClosed(t *testing.T) {
	cache := validCache()
	cache.Predecessor = strings.Repeat("a", 40)
	if err := validateCache(cache, evidenceInput{HeadSHA: strings.Repeat("a", 40)}); err == nil {
		t.Fatal("replayed predecessor was accepted")
	}
}
