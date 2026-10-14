package bodyexecution

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
)

const fieldObservationFile = "gooo-native-field-observation.json"

type CompositionFieldWrite struct {
	SiteIndex        int  `json:"site_index"`
	Frame            int  `json:"frame"`
	UseStart         int  `json:"use_start"`
	UseEnd           int  `json:"use_end"`
	UsesDropped      int  `json:"uses_dropped"`
	Completed        bool `json:"completed"`
	ParentAssignment int  `json:"parent_assignment"`
	Completion       int  `json:"completion"`
}

type CompositionFieldUse struct {
	SiteIndex       int    `json:"site_index"`
	Frame           int    `json:"frame"`
	Assignment      int    `json:"assignment"`
	CompletionClock int    `json:"completion_clock"`
	Reason          string `json:"reason,omitempty"`
}

type CompositionFieldSummary struct {
	Schema          string `json:"schema"`
	Status          string `json:"status"`
	SHA256          string `json:"sha256"`
	MetadataSHA256  string `json:"metadata_sha256"`
	RecordedWrites  int    `json:"recorded_writes"`
	RecordedUses    int    `json:"recorded_uses"`
	Interrupted     int    `json:"interrupted_writes"`
	UnresolvedUses  int    `json:"unresolved_uses"`
	ParentLinks     int    `json:"parent_links"`
	InputBindings   int    `json:"root_input_bindings"`
	SharedRHSWrites int    `json:"shared_rhs_writes"`
	UnknownParents  int    `json:"unknown_parents"`
	Frames          int    `json:"frames"`
	FramesDropped   int    `json:"frames_dropped"`
	WritesDropped   int    `json:"writes_dropped"`
	UsesDropped     int    `json:"uses_dropped"`
	Scope           string `json:"scope"`
}

type compositionFieldRun struct {
	first []byte
	file  nativeFieldFile
}

func (run *compositionFieldRun) reset(root string, r *CompositionRuntime) error {
	if len(r.FieldSites) == 0 {
		return nil
	}
	err := os.Remove(filepath.Join(root, fieldObservationFile))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (run *compositionFieldRun) observe(attempt int, root string, r *CompositionRuntime) error {
	if len(r.FieldSites) == 0 {
		return nil
	}
	file, err := os.Open(filepath.Join(root, fieldObservationFile))
	if err != nil {
		return err
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, (16<<20)+1))
	if err != nil || len(raw) > 16<<20 {
		return fmt.Errorf("native field observation unavailable or exceeds16MiB: %v", err)
	}
	if attempt != 0 {
		if !bytes.Equal(run.first, raw) {
			return fmt.Errorf("native field identities or order differ between runs")
		}
		return nil
	}
	run.first = raw
	return decode(raw, &run.file, 16<<20)
}

func runtimeFieldPlan(r CompositionRuntime) *compositionFieldProjection {
	return &compositionFieldProjection{sourceSHA256: r.FieldProjectionSHA256,
		inventory: compositionFieldInventory{bindings: r.FieldBindings}, sites: r.FieldSites, uses: r.FieldUseSites}
}

func (run *compositionFieldRun) bind(traces []CompositionTrace, r *CompositionRuntime) error {
	if len(r.FieldSites) == 0 {
		return nil
	}
	if err := validateNativeFieldFile(run.file, runtimeFieldPlan(*r), len(traces)); err != nil {
		return err
	}
	for c := range traces {
		traces[c].FieldFrames, traces[c].FieldFrameDropped = run.file.Frames[c], run.file.FrameDropped[c]
		traces[c].FieldWriteDropped, traces[c].FieldUseDropped = run.file.WriteDropped[c], run.file.UseDropped[c]
	}
	for _, row := range run.file.Writes {
		c := row[0]
		traces[c].FieldWrites = append(traces[c].FieldWrites, CompositionFieldWrite{row[1], row[2], row[3], row[4], row[5], row[6] == 1, row[7], row[8]})
	}
	for _, row := range run.file.Uses {
		reason := fieldUseReason(row[3])
		traces[row[0]].FieldUses = append(traces[row[0]].FieldUses, CompositionFieldUse{row[1], row[2], row[3], row[4], reason})
	}
	summary := summarizeFieldObservations(run.file, runtimeFieldPlan(*r))
	summary.SHA256 = digest(run.first)
	r.FieldObservation = summary
	return nil
}

func fieldUseReason(assignment int) string {
	if assignment == -1 {
		return "NO_RECORDED_ASSIGNMENT"
	}
	if assignment == -2 {
		return "WRITE_LIMIT"
	}
	return ""
}

func summarizeFieldObservations(f nativeFieldFile, p *compositionFieldProjection) *CompositionFieldSummary {
	r := &CompositionFieldSummary{Schema: "gooo/native-field-summary/v1", Status: "COMPLETE", RecordedWrites: len(f.Writes), RecordedUses: len(f.Uses),
		MetadataSHA256: compositionDigest([]any{p.sourceSHA256, p.inventory.bindings, p.sites, p.uses}),
		Scope:          "required single scalar record fields; reads link latest completed field writes; explicit copy/call-parameter/known record-return parent links and evaluated field read windows; shared RHS, absent helper parents and scalar-local dependencies remain explicit limits; no values retained"}
	for _, row := range f.Writes {
		site := p.sites[row[1]]
		binding := p.inventory.bindings[site.Binding]
		if row[6] == 0 {
			r.Interrupted++
		}
		if row[7] >= 0 {
			r.ParentLinks++
		} else if site.Kind == "parameter" && binding.InputKind == "root_input" {
			r.InputBindings++
		} else if site.DependencyMode == "field_copy" || site.DependencyMode == "record_return" || site.Kind == "parameter" {
			r.UnknownParents++
		}
		if site.DependencyMode == "shared_rhs" {
			r.SharedRHSWrites++
		}
	}
	for _, row := range f.Uses {
		if row[3] < 0 {
			r.UnresolvedUses++
		}
	}
	for c := range f.Frames {
		r.Frames += f.Frames[c]
		r.FramesDropped += f.FrameDropped[c]
		r.WritesDropped += f.WriteDropped[c]
		r.UsesDropped += f.UseDropped[c]
	}
	if r.Interrupted+r.UnresolvedUses+r.FramesDropped+r.WritesDropped+r.UsesDropped > 0 {
		r.Status = "PARTIAL"
	}
	return r
}

func validateRecordedFieldObservations(r CompositionRuntime) error {
	if r.FieldObservation == nil {
		if len(r.FieldBindings)+len(r.FieldSites)+len(r.FieldUseSites) > 0 || r.FieldProjectionSHA256 != "" {
			return fmt.Errorf("field inventory has no observation summary")
		}
		for _, trace := range r.Traces {
			if len(trace.FieldWrites)+len(trace.FieldUses)+trace.FieldFrames+trace.FieldFrameDropped+trace.FieldWriteDropped+trace.FieldUseDropped != 0 {
				return fmt.Errorf("field history has no observation summary")
			}
		}
		return nil
	}
	p := runtimeFieldPlan(r)
	if err := validateFieldInventory(p); err != nil {
		return err
	}
	f, err := fieldFileFromTraces(r.Traces)
	if err != nil {
		return err
	}
	if err := validateNativeFieldFile(f, p, len(r.Traces)); err != nil {
		return err
	}
	current := summarizeFieldObservations(f, p)
	current.SHA256 = r.FieldObservation.SHA256
	current.Scope = r.FieldObservation.Scope
	if !reflect.DeepEqual(current, r.FieldObservation) {
		return fmt.Errorf("recorded field summary differs from populations or metadata")
	}
	return nil
}
