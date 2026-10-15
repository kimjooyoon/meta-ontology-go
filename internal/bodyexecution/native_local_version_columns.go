package bodyexecution

import (
	"encoding/json"
	"fmt"
)

// LocalVersionColumns retains write identities in one saved local journal.
// The Gooo consumer defines declaration grouping, call grouping and completeness.
type LocalVersionColumns struct {
	SourceSHA256       string   `json:"source_sha256"`
	JournalSHA256      string   `json:"journal_sha256"`
	FileSourcesSHA256  string   `json:"file_sources_sha256"`
	DeclarationScope   string   `json:"declaration_scope"`
	Offset             int      `json:"offset"`
	TotalRows          int      `json:"total_rows"`
	FramesDropped      int      `json:"frames_dropped"`
	WritesDropped      int      `json:"writes_dropped"`
	Cases              []int    `json:"cases"`
	Frames             []int    `json:"frames"`
	Assignments        []int    `json:"assignments"`
	Roots              []string `json:"roots"`
	Activities         []string `json:"activities"`
	DeclarationSHA256  []string `json:"declaration_sha256"`
	DeclarationStarts  []int    `json:"declaration_starts"`
	DeclarationEnds    []int    `json:"declaration_ends"`
	DeclarationMapped  []bool   `json:"declaration_mapped"`
	Packages           []string `json:"declaration_packages"`
	Filenames          []string `json:"declaration_filenames"`
	OriginalActivities []string `json:"declaration_activities"`
	Completed          []bool   `json:"completed"`
}

func SelectRecordedLocalVersions(runtime CompositionRuntime, query NativeOriginQuery) (LocalVersionColumns, *NativeOriginInput, error) {
	var columns LocalVersionColumns
	if query.Kind != "local-writes" {
		return columns, nil, fmt.Errorf("local-version-columns requires local-writes")
	}
	rows, report, err := SelectRecordedNativeOriginPage(runtime, query)
	if err != nil {
		return columns, report, err
	}
	columns = newLocalVersionColumns(runtime, query.Offset, report.TotalRows, len(rows))
	page := nativeOriginPage{}
	page.prepareLocalFileIndices(runtime)
	for i, raw := range rows {
		var fact CompositionLocalWriteFact
		if err := json.Unmarshal(raw, &fact); err != nil {
			return columns, report, err
		}
		columns.appendWrite(i, fact, page.localFileBinding(runtime, fact.SiteIndex, true).Declaration)
	}
	report.InputShape, report.ConsumerRows = "gooo/local-version-columns/v2", 1
	report.Scope = "one page of retained scalar-local write claims and original workspace-file declaration annotations; missing original anchors remain unmapped; original case/frame/assignment numbers are preserved; Gooo owns identity grouping and page status; the population is write events, not complete causal dependencies or all function calls"
	return columns, report, nil
}

func newLocalVersionColumns(runtime CompositionRuntime, offset, total, count int) LocalVersionColumns {
	columns := LocalVersionColumns{SourceSHA256: runtime.SelectedSourceSHA256, JournalSHA256: runtime.LocalObservation.SHA256, DeclarationScope: "original_file",
		Offset: offset, TotalRows: total, FramesDropped: runtime.LocalObservation.FramesDropped, WritesDropped: runtime.LocalObservation.WritesDropped,
		Cases: make([]int, count), Frames: make([]int, count), Assignments: make([]int, count),
		Roots: make([]string, count), Activities: make([]string, count), DeclarationSHA256: make([]string, count),
		DeclarationStarts: make([]int, count), DeclarationEnds: make([]int, count), DeclarationMapped: make([]bool, count), Completed: make([]bool, count),
		Packages: make([]string, count), Filenames: make([]string, count), OriginalActivities: make([]string, count)}
	if runtime.LocalFileSources != nil {
		columns.FileSourcesSHA256 = runtime.LocalFileSources.SHA256
	}
	return columns
}

func (c *LocalVersionColumns) appendWrite(i int, fact CompositionLocalWriteFact, declaration CompositionLocalFileLocation) {
	c.Cases[i], c.Frames[i], c.Assignments[i] = fact.CaseIndex, fact.Frame, fact.AssignmentIndex
	c.Roots[i], c.Activities[i], c.Completed[i] = fact.RootActivityID, fact.ActivityID, fact.Completed
	if declaration.Status == "MAPPED" {
		c.DeclarationSHA256[i], c.DeclarationStarts[i], c.DeclarationEnds[i], c.DeclarationMapped[i] = declaration.SourceSHA256, declaration.Start, declaration.End, true
		c.Packages[i], c.Filenames[i], c.OriginalActivities[i] = declaration.PackagePath, declaration.Filename, declaration.ActivityID
	}
}
