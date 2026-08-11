// Package freshness contains a small executable research contract for
// provenance and projection freshness. It intentionally does not import the
// production semantic, generator, or cache packages.
package freshness

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

const ContractVersion = "freshness-contract/v1"

// Contract is the portable experiment input. Recorded digests may be omitted
// in a fixture; Normalize derives them from the baseline contents.
type Contract struct {
	Version    string   `json:"version"`
	Name       string   `json:"name"`
	Hypothesis string   `json:"hypothesis"`
	Source     Record   `json:"source"`
	Records    []Record `json:"records"`
	Required   []string `json:"required"`
	Cases      []Case   `json:"cases"`
}

// Record describes one source, generated projection, cache projection, or
// evidence entity. InputIDs are dependency edges; UsedIDs are PROV edges.
type Record struct {
	ID                    string     `json:"id"`
	Kind                  string     `json:"kind"`
	Content               string     `json:"content"`
	InputIDs              []string   `json:"input_ids,omitempty"`
	RecordedInputDigest   string     `json:"recorded_input_digest,omitempty"`
	RecordedContentDigest string     `json:"recorded_content_digest,omitempty"`
	Provenance            Provenance `json:"provenance"`
}

// Provenance is the minimal PROV-shaped relation envelope used by the
// experiment. Activity IDs remain opaque to this package.
type Provenance struct {
	ActivityID string   `json:"activity_id"`
	EntityID   string   `json:"entity_id"`
	UsedIDs    []string `json:"used_ids,omitempty"`
}

// Case applies one controlled mutation to the baseline fixture.
type Case struct {
	ID              string        `json:"id"`
	Mutation        Mutation      `json:"mutation"`
	ExpectedVerdict string        `json:"expected_verdict"`
	Expected        []Expectation `json:"expected,omitempty"`
	Reason          string        `json:"reason,omitempty"`
}

// Mutation is deliberately finite so a future implementation can reproduce
// the same cases from an AST/IR or filesystem adapter.
type Mutation struct {
	Kind     string   `json:"kind"`
	RecordID string   `json:"record_id,omitempty"`
	Content  string   `json:"content,omitempty"`
	RemoveID string   `json:"remove_id,omitempty"`
	UsedIDs  []string `json:"used_ids,omitempty"`
}

// Expectation names the observable state required for one record.
type Expectation struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	Status Status `json:"status"`
}

// Status is the observable record state produced by the reference evaluator.
type Status string

const (
	StatusFresh   Status = "fresh"
	StatusStale   Status = "stale"
	StatusMissing Status = "missing"
	StatusInvalid Status = "invalid"
)

// Verdict distinguishes demonstrated behavior from intentionally unimplemented
// future behavior. Deferred is never treated as a pass.
type Verdict string

const (
	VerdictPass     Verdict = "pass"
	VerdictFail     Verdict = "fail"
	VerdictDeferred Verdict = "deferred"
)

// Load reads and normalizes a JSON experiment contract.
func Load(path string) (Contract, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Contract{}, fmt.Errorf("read freshness contract: %w", err)
	}
	var contract Contract
	if err := json.Unmarshal(data, &contract); err != nil {
		return Contract{}, fmt.Errorf("decode freshness contract: %w", err)
	}
	if err := contract.Normalize(); err != nil {
		return Contract{}, err
	}
	return contract, nil
}

// Normalize validates identity and dependency boundaries, sorts order-only
// fields, and derives omitted baseline digests.
func (c *Contract) Normalize() error {
	if c.Version != ContractVersion {
		return fmt.Errorf("unsupported freshness contract version %q", c.Version)
	}
	if c.Source.ID == "" || c.Source.Content == "" {
		return fmt.Errorf("source ID and content are required")
	}
	if c.Source.Kind != "source" {
		return fmt.Errorf("source kind must be source")
	}
	known := map[string]Record{c.Source.ID: c.Source}
	for i := range c.Records {
		record := &c.Records[i]
		if record.ID == "" || record.Kind == "" || record.Content == "" {
			return fmt.Errorf("record %d has incomplete identity or content", i)
		}
		switch record.Kind {
		case "generated-projection", "cache", "evidence":
		default:
			return fmt.Errorf("record %q has unsupported kind %q", record.ID, record.Kind)
		}
		if _, exists := known[record.ID]; exists {
			return fmt.Errorf("duplicate record ID %q", record.ID)
		}
		known[record.ID] = *record
	}
	if err := c.normalizeRecords(known); err != nil {
		return err
	}
	if err := validateCases(c, known); err != nil {
		return err
	}
	if len(c.Cases) == 0 {
		return fmt.Errorf("at least one experiment case is required")
	}
	for _, id := range c.Required {
		if _, ok := known[id]; !ok {
			return fmt.Errorf("required ID %q is unknown", id)
		}
	}
	for i := range c.Cases {
		sort.Strings(c.Cases[i].Mutation.UsedIDs)
		sort.SliceStable(c.Cases[i].Expected, func(left, right int) bool {
			if c.Cases[i].Expected[left].Kind != c.Cases[i].Expected[right].Kind {
				return c.Cases[i].Expected[left].Kind < c.Cases[i].Expected[right].Kind
			}
			return c.Cases[i].Expected[left].ID < c.Cases[i].Expected[right].ID
		})
	}
	sort.Strings(c.Required)
	sort.SliceStable(c.Records, func(i, j int) bool { return c.Records[i].ID < c.Records[j].ID })
	sort.SliceStable(c.Cases, func(i, j int) bool { return c.Cases[i].ID < c.Cases[j].ID })
	return nil
}

func (c *Contract) normalizeRecords(known map[string]Record) error {
	for i := range c.Records {
		record := &c.Records[i]
		sort.Strings(record.InputIDs)
		sort.Strings(record.Provenance.UsedIDs)
		if err := validateReferences(*record, known); err != nil {
			return err
		}
		contentDigest := hashString(record.Content)
		inputDigest, err := digestInputs(record.InputIDs, known)
		if err != nil {
			return fmt.Errorf("record %q: %w", record.ID, err)
		}
		if record.RecordedContentDigest == "" {
			record.RecordedContentDigest = contentDigest
		} else if record.RecordedContentDigest != contentDigest {
			return fmt.Errorf("record %q has incorrect recorded content digest", record.ID)
		}
		if record.RecordedInputDigest == "" {
			record.RecordedInputDigest = inputDigest
		} else if record.RecordedInputDigest != inputDigest {
			return fmt.Errorf("record %q has incorrect recorded input digest", record.ID)
		}
	}
	return nil
}

func validateReferences(record Record, known map[string]Record) error {
	if record.Provenance.ActivityID == "" || record.Provenance.EntityID == "" {
		return fmt.Errorf("record %q has incomplete provenance", record.ID)
	}
	for _, id := range append(append([]string{}, record.InputIDs...), record.Provenance.UsedIDs...) {
		if _, ok := known[id]; !ok {
			return fmt.Errorf("record %q references unknown ID %q", record.ID, id)
		}
	}
	return nil
}

func validateCases(c *Contract, known map[string]Record) error {
	caseIDs := make(map[string]struct{}, len(c.Cases))
	for _, testCase := range c.Cases {
		if testCase.ID == "" || testCase.ExpectedVerdict == "" {
			return fmt.Errorf("experiment case has incomplete identity or verdict")
		}
		if _, exists := caseIDs[testCase.ID]; exists {
			return fmt.Errorf("duplicate experiment case ID %q", testCase.ID)
		}
		caseIDs[testCase.ID] = struct{}{}
		if testCase.ExpectedVerdict != string(VerdictPass) && testCase.ExpectedVerdict != string(VerdictDeferred) {
			return fmt.Errorf("case %q has unsupported expected verdict", testCase.ID)
		}
		switch testCase.Mutation.Kind {
		case "none", "content", "remove", "used-ids", "future-self-hosted":
		default:
			return fmt.Errorf("case %q has unsupported mutation kind", testCase.ID)
		}
		if err := validateMutation(testCase); err != nil {
			return err
		}
		if testCase.Mutation.RecordID != "" {
			if _, ok := known[testCase.Mutation.RecordID]; !ok {
				return fmt.Errorf("case %q mutates unknown ID %q", testCase.ID, testCase.Mutation.RecordID)
			}
		}
		if testCase.Mutation.RemoveID != "" {
			if _, ok := known[testCase.Mutation.RemoveID]; !ok {
				return fmt.Errorf("case %q removes unknown ID %q", testCase.ID, testCase.Mutation.RemoveID)
			}
		}
		for _, expected := range testCase.Expected {
			if expected.Status != StatusFresh && expected.Status != StatusStale && expected.Status != StatusMissing && expected.Status != StatusInvalid {
				return fmt.Errorf("case %q has unsupported expected status", testCase.ID)
			}
			if _, ok := known[expected.ID]; !ok && expected.Status != StatusMissing {
				return fmt.Errorf("case %q expects unknown ID %q", testCase.ID, expected.ID)
			}
		}
	}
	return nil
}

func validateMutation(testCase Case) error {
	if testCase.Mutation.Kind == "future-self-hosted" && testCase.ExpectedVerdict != string(VerdictDeferred) {
		return fmt.Errorf("case %q must defer the future self-hosted mutation", testCase.ID)
	}
	if testCase.ExpectedVerdict == string(VerdictDeferred) && testCase.Mutation.Kind != "future-self-hosted" {
		return fmt.Errorf("case %q defers an implemented mutation", testCase.ID)
	}
	switch testCase.Mutation.Kind {
	case "content":
		if testCase.Mutation.RecordID == "" || testCase.Mutation.Content == "" {
			return fmt.Errorf("case %q content mutation is incomplete", testCase.ID)
		}
	case "remove":
		if testCase.Mutation.RemoveID == "" {
			return fmt.Errorf("case %q remove mutation has no ID", testCase.ID)
		}
	case "used-ids":
		if testCase.Mutation.RecordID == "" {
			return fmt.Errorf("case %q provenance mutation has no record ID", testCase.ID)
		}
	}
	return nil
}

func digestInputs(ids []string, records map[string]Record) (string, error) {
	ordered := append([]string(nil), ids...)
	sort.Strings(ordered)
	buffer := make([]byte, 0, len(ordered)*80)
	for _, id := range ordered {
		record, ok := records[id]
		if !ok {
			return "", fmt.Errorf("input %q is unavailable", id)
		}
		buffer = append(buffer, id...)
		buffer = append(buffer, 0)
		buffer = append(buffer, hashString(record.Content)...)
		buffer = append(buffer, '\n')
	}
	return hashBytes(buffer), nil
}

func hashString(value string) string { return hashBytes([]byte(value)) }

func hashBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}
