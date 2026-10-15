package bodyexecution

import (
	"encoding/json"
	"fmt"
)

// LocalUseColumns transports one bounded page and the referenced write facts.
// The Gooo reader owns the generation/usage relation and unresolved status.
type LocalUseColumns struct {
	ReceiptSHA256     string   `json:"receipt_sha256"`
	SourceSHA256      string   `json:"source_sha256"`
	JournalSHA256     string   `json:"journal_sha256"`
	Offset            int      `json:"offset"`
	TotalRows         int      `json:"total_rows"`
	FramesDropped     int      `json:"frames_dropped"`
	WritesDropped     int      `json:"writes_dropped"`
	UsesDropped       int      `json:"uses_dropped"`
	Cases             []int    `json:"cases"`
	UseIndices        []int    `json:"use_indices"`
	Frames            []int    `json:"frames"`
	Assignments       []int    `json:"assignments"`
	Roots             []string `json:"roots"`
	Activities        []string `json:"activities"`
	Resolved          []bool   `json:"resolved"`
	Reasons           []string `json:"reasons"`
	VersionFrames     []int    `json:"version_frames"`
	VersionActivities []string `json:"version_activities"`
	VersionCompleted  []bool   `json:"version_completed"`
	ReadSHA256        []string `json:"read_sha256"`
	ReadStarts        []int    `json:"read_starts"`
	ReadEnds          []int    `json:"read_ends"`
	ReadMappings      []string `json:"read_mappings"`
	Names             []string `json:"names"`
	Ordinals          []int    `json:"ordinals"`
	ScalarTypes       []string `json:"scalar_types"`
}

func SelectRecordedLocalUses(runtime CompositionRuntime, query NativeOriginQuery, receiptSHA256 string) (LocalUseColumns, *NativeOriginInput, error) {
	var columns LocalUseColumns
	if query.Kind != "local-uses" {
		return columns, nil, fmt.Errorf("local-use-columns requires local-uses")
	}
	rows, report, err := SelectRecordedNativeOriginPage(runtime, query)
	if err != nil {
		return columns, report, err
	}
	indices, err := localUseTraceIndices(runtime.Traces)
	if err != nil {
		return columns, report, err
	}
	columns = newLocalUseColumns(runtime, query.Offset, report.TotalRows, len(rows), receiptSHA256)
	for i, raw := range rows {
		var fact CompositionLocalUseFact
		if err := json.Unmarshal(raw, &fact); err != nil {
			return columns, report, err
		}
		index, found := indices[fact.CaseIndex]
		if !found {
			return columns, report, fmt.Errorf("local use case has no recorded trace")
		}
		columns.appendUse(i, fact, runtime.Traces[index], runtime.LocalSites)
	}
	report.InputShape, report.ConsumerRows = "gooo/local-use-columns/v1", 1
	report.Scope = "one page of recorded local reads and their referenced write facts; original case/use/frame/assignment numbers retained; receipt digest names saved JSON bytes; Gooo owns generation/usage relations and unresolved status; read locations are selected-source coordinates"
	return columns, report, nil
}

func localUseTraceIndices(traces []CompositionTrace) (map[int]int, error) {
	indices := make(map[int]int, len(traces))
	for i := range traces {
		if _, exists := indices[traces[i].CaseIndex]; exists {
			return nil, fmt.Errorf("recorded local use case identity is duplicated")
		}
		indices[traces[i].CaseIndex] = i
	}
	return indices, nil
}

func newLocalUseColumns(runtime CompositionRuntime, offset, total, count int, receiptSHA256 string) LocalUseColumns {
	return LocalUseColumns{ReceiptSHA256: receiptSHA256, SourceSHA256: runtime.SelectedSourceSHA256, JournalSHA256: runtime.LocalObservation.SHA256,
		Offset: offset, TotalRows: total, FramesDropped: runtime.LocalObservation.FramesDropped, WritesDropped: runtime.LocalObservation.WritesDropped, UsesDropped: runtime.LocalObservation.UsesDropped,
		Cases: make([]int, count), UseIndices: make([]int, count), Frames: make([]int, count), Assignments: make([]int, count),
		Roots: make([]string, count), Activities: make([]string, count), Resolved: make([]bool, count), Reasons: make([]string, count),
		VersionFrames: make([]int, count), VersionActivities: make([]string, count), VersionCompleted: make([]bool, count),
		ReadSHA256: make([]string, count), ReadStarts: make([]int, count), ReadEnds: make([]int, count), ReadMappings: make([]string, count),
		Names: make([]string, count), Ordinals: make([]int, count), ScalarTypes: make([]string, count)}
}

func (c *LocalUseColumns) appendUse(i int, fact CompositionLocalUseFact, trace CompositionTrace, sites []CompositionLocalSite) {
	c.Cases[i], c.UseIndices[i], c.Frames[i], c.Assignments[i] = fact.CaseIndex, fact.UseIndex, fact.Frame, fact.AssignmentIndex
	c.Roots[i], c.Activities[i], c.Resolved[i], c.Reasons[i] = fact.RootActivityID, fact.ActivityID, fact.Resolved, fact.Reason
	c.ReadSHA256[i], c.ReadStarts[i], c.ReadEnds[i], c.ReadMappings[i] = fact.SourceSHA256, fact.SourceStart, fact.SourceEnd, fact.SourceMapping
	c.Names[i], c.Ordinals[i], c.ScalarTypes[i] = fact.VariableName, fact.VariableOrdinal, fact.ScalarType
	c.VersionFrames[i] = -1
	if fact.AssignmentIndex >= 0 {
		write := trace.LocalWrites[fact.AssignmentIndex]
		c.VersionFrames[i], c.VersionCompleted[i] = write.Frame, write.Completed
		c.VersionActivities[i] = sites[write.SiteIndex].ActivityID
	}
}
