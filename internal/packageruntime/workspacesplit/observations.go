package workspacesplit

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

var ErrSourceChanged = errors.New("source changed between workspace observations")

type sourceObservations struct {
	reader   workspaceinput.Reader
	consumer string
	values   map[string]InputObservation
	bytes    int
}

func (r *sourceObservations) ReadFile(path string) ([]byte, error) {
	if r.reader == nil {
		return nil, fmt.Errorf("workspace source reader is required")
	}
	raw, err := r.reader.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(raw) > workspaceinput.MaxSourceBytes {
		return nil, fmt.Errorf("source exceeds the workspace byte limit")
	}
	address, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	digest := fmt.Sprintf("sha256:%x", sha256.Sum256(raw))
	previous, seen := r.values[address]
	if seen && (previous.SHA256 != digest || previous.ByteLength != len(raw)) {
		return nil, fmt.Errorf("%w: %s", ErrSourceChanged, address)
	}
	if !seen {
		r.bytes += len(raw)
		if r.bytes > workspaceinput.MaxSourceSetBytes {
			return nil, fmt.Errorf("observed source set exceeds the workspace byte limit")
		}
		previous = InputObservation{Path: address, SHA256: digest, ByteLength: len(raw)}
	}
	if !slices.Contains(previous.Consumers, r.consumer) {
		previous.Consumers = append(previous.Consumers, r.consumer)
	}
	r.values[address] = previous
	return raw, nil
}

func (r *sourceObservations) list() []InputObservation {
	keys := make([]string, 0, len(r.values))
	for key := range r.values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	result := make([]InputObservation, 0, len(keys))
	for _, key := range keys {
		value := r.values[key]
		value.Consumers = slices.Clone(value.Consumers)
		slices.Sort(value.Consumers)
		result = append(result, value)
	}
	return result
}
