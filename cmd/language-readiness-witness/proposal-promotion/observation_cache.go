package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy/proposalpredecessor"
)

const proposalObservationCacheSchema = proposalpredecessor.ObservationSchema

// This cache contains only raw HTTP observations. It deliberately has no
// selection, conclusion, or promotion fields.
type proposalObservedResponse struct {
	Kind       string `json:"kind"`
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	Body       []byte `json:"body"`
	Link       string `json:"link,omitempty"`
	Location   string `json:"location,omitempty"`
	Failure    string `json:"failure,omitempty"`
}

type proposalObservationCacheFile struct {
	Schema    string                     `json:"schema"`
	Responses []proposalObservedResponse `json:"responses"`
}

type proposalObservationStore struct {
	path         string
	replay       bool
	responses    []proposalObservedResponse
	position     int
	consumed     int
	canonicalRaw []byte
	cacheDigest  string
}

func openProposalObservationStore(capture, replay string) (*proposalObservationStore, error) {
	if capture != "" && replay != "" {
		return nil, fmt.Errorf("proposal observation capture and replay are mutually exclusive")
	}
	if replay == "" {
		if capture == "" {
			return nil, nil
		}
		return &proposalObservationStore{path: capture}, nil
	}
	raw, err := os.ReadFile(replay)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var file proposalObservationCacheFile
	if err := decoder.Decode(&file); err != nil {
		return nil, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, fmt.Errorf("proposal observation cache has trailing content")
	}
	if file.Schema != proposalObservationCacheSchema {
		return nil, fmt.Errorf("proposal observation cache schema mismatch")
	}
	return &proposalObservationStore{
		path: replay, replay: true, responses: file.Responses,
		canonicalRaw: append([]byte(nil), raw...), cacheDigest: proposalObservationDigest(raw),
	}, nil
}

func (store *proposalObservationStore) record(response proposalObservedResponse) {
	if store == nil || store.replay {
		return
	}
	store.responses = append(store.responses, response)
	store.consumed++
}

type proposalObservationReplayError struct{ Err error }

func (failure *proposalObservationReplayError) Error() string { return failure.Err.Error() }
func (failure *proposalObservationReplayError) Unwrap() error { return failure.Err }

func (store *proposalObservationStore) next(kind, targetURL string) (proposalObservedResponse, error) {
	if store == nil || !store.replay {
		return proposalObservedResponse{}, &proposalObservationReplayError{Err: fmt.Errorf("proposal observation replay is not enabled")}
	}
	if store.position >= len(store.responses) {
		return proposalObservedResponse{}, &proposalObservationReplayError{Err: fmt.Errorf("proposal observation replay exhausted at %s %s", kind, targetURL)}
	}
	response := store.responses[store.position]
	store.position++
	if response.Kind != kind || response.URL != targetURL {
		return proposalObservedResponse{}, &proposalObservationReplayError{Err: fmt.Errorf("proposal observation replay coordinate mismatch: got %s %s want %s %s", response.Kind, response.URL, kind, targetURL)}
	}
	return response, nil
}

func (store *proposalObservationStore) finalize() (proposalpredecessor.ObservationEvidence, error) {
	if store == nil {
		return proposalpredecessor.ObservationEvidence{}, fmt.Errorf("proposal observation evidence is unavailable")
	}
	if store.replay {
		if store.position != len(store.responses) {
			return proposalpredecessor.ObservationEvidence{}, &proposalObservationReplayError{Err: fmt.Errorf("proposal observation replay left %d responses unused", len(store.responses)-store.position)}
		}
		currentRaw, err := os.ReadFile(store.path)
		if err != nil {
			return proposalpredecessor.ObservationEvidence{}, err
		}
		if !bytes.Equal(currentRaw, store.canonicalRaw) {
			return proposalpredecessor.ObservationEvidence{}, fmt.Errorf("proposal observation replay cache changed after first read")
		}
		return proposalObservationEvidence(store.path, store.canonicalRaw, store.cacheDigest, len(store.responses), store.position)
	}
	file := proposalObservationCacheFile{Schema: proposalObservationCacheSchema, Responses: store.responses}
	raw, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return proposalpredecessor.ObservationEvidence{}, err
	}
	raw = append(raw, '\n')
	if err := os.MkdirAll(filepath.Dir(store.path), 0o755); err != nil {
		return proposalpredecessor.ObservationEvidence{}, err
	}
	if err := os.WriteFile(store.path, raw, 0o600); err != nil {
		return proposalpredecessor.ObservationEvidence{}, err
	}
	canonicalRaw, err := os.ReadFile(store.path)
	if err != nil {
		return proposalpredecessor.ObservationEvidence{}, err
	}
	if !bytes.Equal(canonicalRaw, raw) {
		return proposalpredecessor.ObservationEvidence{}, fmt.Errorf("proposal observation cache changed during finalize")
	}
	store.canonicalRaw = canonicalRaw
	store.cacheDigest = proposalObservationDigest(canonicalRaw)
	return proposalObservationEvidence(store.path, canonicalRaw, store.cacheDigest, len(store.responses), store.consumed)
}

func proposalObservationEvidence(path string, raw []byte, digest string, total, consumed int) (proposalpredecessor.ObservationEvidence, error) {
	if proposalObservationDigest(raw) != digest {
		return proposalpredecessor.ObservationEvidence{}, fmt.Errorf("proposal observation cache digest changed during finalize")
	}
	evidence := proposalpredecessor.ObservationEvidence{
		Schema: proposalObservationCacheSchema, CachePath: path, CacheBytes: len(raw),
		CacheDigest: digest, ResponseTotal: total, ResponseConsumed: consumed,
	}
	if err := proposalpredecessor.ValidateObservationEvidence(evidence); err != nil {
		return proposalpredecessor.ObservationEvidence{}, err
	}
	return evidence, nil
}

func proposalObservationDigest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}
