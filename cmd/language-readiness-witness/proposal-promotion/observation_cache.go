package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const proposalObservationCacheSchema = "gooo/language-readiness-api-observation/v1"

// This cache contains only raw HTTP observations. It deliberately has no
// selection, conclusion, or promotion fields.
type proposalObservedResponse struct {
	Kind       string `json:"kind"`
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	Body       []byte `json:"body"`
	Link       string `json:"link,omitempty"`
	Failure    string `json:"failure,omitempty"`
}

type proposalObservationCacheFile struct {
	Schema    string                     `json:"schema"`
	Responses []proposalObservedResponse `json:"responses"`
}

type proposalObservationStore struct {
	path      string
	replay    bool
	responses []proposalObservedResponse
	position  int
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
	return &proposalObservationStore{path: replay, replay: true, responses: file.Responses}, nil
}

func (store *proposalObservationStore) record(response proposalObservedResponse) {
	if store == nil || store.replay {
		return
	}
	store.responses = append(store.responses, response)
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

func (store *proposalObservationStore) close() error {
	if store == nil {
		return nil
	}
	if store.replay {
		if store.position != len(store.responses) {
			return &proposalObservationReplayError{Err: fmt.Errorf("proposal observation replay left %d responses unused", len(store.responses)-store.position)}
		}
		return nil
	}
	file := proposalObservationCacheFile{Schema: proposalObservationCacheSchema, Responses: store.responses}
	raw, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	if err := os.MkdirAll(filepath.Dir(store.path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(store.path, raw, 0o600)
}
