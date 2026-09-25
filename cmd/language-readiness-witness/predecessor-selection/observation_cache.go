package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const observationCacheSchema = "gooo/language-readiness-api-observation/v1"

type observedResponse struct {
	Kind       string `json:"kind"`
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	Body       []byte `json:"body"`
	Link       string `json:"link,omitempty"`
	Failure    string `json:"failure,omitempty"`
}

type observationCacheFile struct {
	Schema    string             `json:"schema"`
	Responses []observedResponse `json:"responses"`
}

type observationStore struct {
	path      string
	replay    bool
	responses []observedResponse
	position  int
}

func openObservationStore(capture, replay string) (*observationStore, error) {
	if capture != "" && replay != "" {
		return nil, fmt.Errorf("observation capture and replay are mutually exclusive")
	}
	if replay == "" {
		if capture == "" {
			return nil, nil
		}
		return &observationStore{path: capture}, nil
	}
	raw, err := os.ReadFile(replay)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var file observationCacheFile
	if err := decoder.Decode(&file); err != nil {
		return nil, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, fmt.Errorf("observation cache has trailing content")
	}
	if file.Schema != observationCacheSchema {
		return nil, fmt.Errorf("observation cache schema mismatch")
	}
	return &observationStore{path: replay, replay: true,
		responses: file.Responses}, nil
}

func (store *observationStore) record(response observedResponse) {
	if store == nil || store.replay {
		return
	}
	store.responses = append(store.responses, response)
}

func (store *observationStore) next(kind, url string) (observedResponse, error) {
	if store == nil || !store.replay {
		return observedResponse{}, fmt.Errorf("observation replay is not enabled")
	}
	if store.position >= len(store.responses) {
		return observedResponse{}, fmt.Errorf("observation replay exhausted at %s %s", kind, url)
	}
	response := store.responses[store.position]
	store.position++
	if response.Kind != kind || response.URL != url {
		return observedResponse{}, fmt.Errorf("observation replay coordinate mismatch: got %s %s want %s %s", response.Kind, response.URL, kind, url)
	}
	return response, nil
}

func (store *observationStore) close() error {
	if store == nil {
		return nil
	}
	if store.replay {
		if store.position != len(store.responses) {
			return fmt.Errorf("observation replay left %d responses unused", len(store.responses)-store.position)
		}
		return nil
	}
	file := observationCacheFile{Schema: observationCacheSchema,
		Responses: store.responses}
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
