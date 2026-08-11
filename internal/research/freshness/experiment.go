package freshness

import (
	"encoding/json"
	"fmt"
	"sort"
)

// Report is the deterministic output of one contract run.
type Report struct {
	Name       string       `json:"name"`
	Hypothesis string       `json:"hypothesis"`
	Results    []CaseResult `json:"results"`
}

// CaseResult separates measured observations from the declared expectation.
type CaseResult struct {
	ID              string        `json:"id"`
	ExpectedVerdict string        `json:"expected_verdict"`
	Verdict         Verdict       `json:"verdict"`
	Expected        []Expectation `json:"expected,omitempty"`
	Observed        []Observation `json:"observed,omitempty"`
	Measurement     Measurement   `json:"measurement"`
	Detail          string        `json:"detail,omitempty"`
}

// Observation is one materialized state in a case.
type Observation struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	Status Status `json:"status"`
}

// Measurement contains reproducible structural values, not wall-clock time.
type Measurement struct {
	ActiveRecords   int    `json:"active_records"`
	DependencyEdges int    `json:"dependency_edges"`
	ProvenanceEdges int    `json:"provenance_edges"`
	ContentBytes    int    `json:"content_bytes"`
	NonFresh        int    `json:"non_fresh"`
	Fingerprint     string `json:"fingerprint"`
}

// Run evaluates all cases. A failed case is data in the report, not a Go
// error; errors are reserved for malformed contracts.
func Run(contract Contract) (Report, error) {
	if err := contract.Normalize(); err != nil {
		return Report{}, err
	}
	report := Report{Name: contract.Name, Hypothesis: contract.Hypothesis, Results: make([]CaseResult, 0, len(contract.Cases))}
	for _, testCase := range contract.Cases {
		result := runCase(contract, testCase)
		report.Results = append(report.Results, result)
	}
	sort.SliceStable(report.Results, func(i, j int) bool { return report.Results[i].ID < report.Results[j].ID })
	return report, nil
}

// Passed is true only when every case has demonstrated pass. Deferred cases
// therefore keep the experiment explicitly incomplete.
func (r Report) Passed() bool {
	if len(r.Results) == 0 {
		return false
	}
	for _, result := range r.Results {
		if result.Verdict != VerdictPass {
			return false
		}
	}
	return true
}

// HasFailures distinguishes a failed hypothesis from an intentionally
// deferred future capability.
func (r Report) HasFailures() bool {
	for _, result := range r.Results {
		if result.Verdict == VerdictFail {
			return true
		}
	}
	return false
}

// HasDeferred reports whether any case still needs a future implementation.
func (r Report) HasDeferred() bool {
	for _, result := range r.Results {
		if result.Verdict == VerdictDeferred {
			return true
		}
	}
	return false
}

// CanonicalJSON returns normalized contract bytes suitable for evidence and
// cache keys. JSON field order is fixed by the Go structs after normalization.
func CanonicalJSON(contract Contract) ([]byte, error) {
	if err := contract.Normalize(); err != nil {
		return nil, err
	}
	return json.Marshal(contract)
}

func runCase(contract Contract, testCase Case) CaseResult {
	result := CaseResult{ID: testCase.ID, ExpectedVerdict: testCase.ExpectedVerdict, Expected: testCase.Expected}
	if testCase.ExpectedVerdict == string(VerdictDeferred) {
		result.Verdict = VerdictDeferred
		result.Detail = testCase.Reason
		result.Measurement = measure(contract, nil, nil)
		return result
	}
	observed, active := evaluate(contract, testCase.Mutation)
	result.Observed = observed
	result.Measurement = measure(contract, active, observed)
	result.Verdict = VerdictPass
	if !matches(observed, testCase.Expected) {
		result.Verdict = VerdictFail
		result.Detail = "observed states do not match expectations"
	}
	return result
}

func evaluate(contract Contract, mutation Mutation) ([]Observation, map[string]Record) {
	active := make(map[string]Record, len(contract.Records)+1)
	active[contract.Source.ID] = contract.Source
	for _, record := range contract.Records {
		active[record.ID] = record
	}
	applyMutation(active, mutation)
	current := make(map[string]Record, len(active))
	for id, record := range active {
		current[id] = record
	}
	states := make(map[string]Status, len(active))
	visiting := make(map[string]bool, len(active))
	ids := sortedRecordIDs(active)
	observed := make([]Observation, 0, len(ids)+len(contract.Required))
	for _, id := range ids {
		status := evaluateRecord(id, active, current, states, visiting)
		observed = append(observed, Observation{ID: id, Kind: active[id].Kind, Status: status})
	}
	for _, id := range contract.Required {
		if _, ok := active[id]; !ok {
			observed = append(observed, Observation{ID: id, Kind: requiredKind(contract, id), Status: StatusMissing})
		}
	}
	sort.SliceStable(observed, func(i, j int) bool { return observed[i].ID < observed[j].ID })
	return observed, active
}

func applyMutation(active map[string]Record, mutation Mutation) {
	switch mutation.Kind {
	case "content":
		record := active[mutation.RecordID]
		record.Content = mutation.Content
		active[mutation.RecordID] = record
	case "remove":
		delete(active, mutation.RemoveID)
	case "used-ids":
		record := active[mutation.RecordID]
		record.Provenance.UsedIDs = append([]string(nil), mutation.UsedIDs...)
		active[mutation.RecordID] = record
	}
}

func evaluateRecord(id string, active, current map[string]Record, states map[string]Status, visiting map[string]bool) Status {
	if status, ok := states[id]; ok {
		return status
	}
	record, ok := active[id]
	if !ok {
		return StatusMissing
	}
	if id == "" || record.ID != id {
		return StatusInvalid
	}
	if record.Kind == "source" {
		states[id] = StatusFresh
		return StatusFresh
	}
	if visiting[id] {
		states[id] = StatusInvalid
		return StatusInvalid
	}
	visiting[id] = true
	status := StatusFresh
	for _, inputID := range record.InputIDs {
		dependency := evaluateRecord(inputID, active, current, states, visiting)
		if dependency != StatusFresh {
			status = StatusStale
		}
	}
	if status == StatusFresh {
		if err := validateUsedIDs(record, active); err != nil {
			status = StatusInvalid
		} else if inputDigest, err := digestInputs(record.InputIDs, current); err != nil || inputDigest != record.RecordedInputDigest {
			status = StatusStale
		} else if hashString(record.Content) != record.RecordedContentDigest {
			status = StatusStale
		}
	}
	delete(visiting, id)
	states[id] = status
	return status
}

func validateUsedIDs(record Record, active map[string]Record) error {
	for _, id := range record.Provenance.UsedIDs {
		if id == record.ID {
			return fmt.Errorf("record uses itself")
		}
		if _, ok := active[id]; !ok {
			return fmt.Errorf("record uses missing ID")
		}
	}
	return nil
}

func sortedRecordIDs(records map[string]Record) []string {
	ids := make([]string, 0, len(records))
	for id := range records {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func requiredKind(contract Contract, id string) string {
	if id == contract.Source.ID {
		return contract.Source.Kind
	}
	for _, record := range contract.Records {
		if record.ID == id {
			return record.Kind
		}
	}
	return "unknown"
}

func matches(observed []Observation, expected []Expectation) bool {
	states := make(map[string]Status, len(observed))
	for _, item := range observed {
		states[item.Kind+"\x00"+item.ID] = item.Status
	}
	for _, item := range expected {
		if states[item.Kind+"\x00"+item.ID] != item.Status {
			return false
		}
	}
	return true
}

func measure(contract Contract, active map[string]Record, observed []Observation) Measurement {
	measurement := Measurement{}
	if active == nil {
		active = map[string]Record{contract.Source.ID: contract.Source}
		for _, record := range contract.Records {
			active[record.ID] = record
		}
	}
	for _, record := range active {
		measurement.ActiveRecords++
		measurement.ContentBytes += len(record.Content)
		measurement.DependencyEdges += len(record.InputIDs)
		measurement.ProvenanceEdges += len(record.Provenance.UsedIDs)
	}
	for _, item := range observed {
		if item.Status != StatusFresh {
			measurement.NonFresh++
		}
	}
	payload := struct {
		ActiveRecords int
		Observed      []Observation
	}{measurement.ActiveRecords, observed}
	data, _ := json.Marshal(payload)
	measurement.Fingerprint = hashBytes(data)
	return measurement
}
