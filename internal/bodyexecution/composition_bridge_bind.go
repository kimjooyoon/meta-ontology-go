package bodyexecution

import (
	"fmt"
	"reflect"
	"strings"
)

func (run *compositionBridgeRun) bind(traces []CompositionTrace, r *CompositionRuntime) error {
	if len(r.FieldSites) == 0 {
		return nil
	}
	if err := bindBridgeFile(run.file, traces); err != nil {
		return err
	}
	summary, err := validateBridgeObservations(traces, *r)
	if err != nil {
		return err
	}
	summary.SHA256 = digest(run.first)
	r.BridgeObservation = summary
	return nil
}

func bindBridgeFile(f nativeBridgeFile, traces []CompositionTrace) error {
	if f.Schema != "gooo/native-field-local-bridge/v1" || len(traces) < 1 || len(traces) > 128 || len(f.Completions) != len(traces) || len(f.Locals) > 4096 || len(f.Fields) > 4096 || len(f.LocalClocks) > 65536 || len(f.FieldClocks) > 65536 {
		return fmt.Errorf("native field/local bridge schema or populations differ")
	}
	local, field := 0, 0
	for c := range traces {
		t := &traces[c]
		if t.CaseIndex != c || local+len(t.LocalUses) > len(f.LocalClocks) || field+len(t.FieldUses) > len(f.FieldClocks) {
			return fmt.Errorf("native bridge case or read-clock population differs")
		}
		t.Bridge = &CompositionBridgeTrace{Completions: f.Completions[c], LocalClocks: append([]int(nil), f.LocalClocks[local:local+len(t.LocalUses)]...), FieldClocks: append([]int(nil), f.FieldClocks[field:field+len(t.FieldUses)]...)}
		local, field = local+len(t.LocalUses), field+len(t.FieldUses)
	}
	if local != len(f.LocalClocks) || field != len(f.FieldClocks) {
		return fmt.Errorf("native bridge has extra read clocks")
	}
	if err := bindBridgeWindows(f.Locals, traces, false); err != nil {
		return err
	}
	return bindBridgeWindows(f.Fields, traces, true)
}

func bindBridgeWindows(rows [][]int, traces []CompositionTrace, field bool) error {
	last := -1
	for _, row := range rows {
		if len(row) != 7 || row[0] < 0 || row[0] >= len(traces) || row[0] < last {
			return fmt.Errorf("native bridge window width or case order differs")
		}
		last = row[0]
		windows := &traces[last].Bridge.Locals
		if field {
			windows = &traces[last].Bridge.Fields
		}
		if row[1] != len(*windows) {
			return fmt.Errorf("native bridge assignment index differs")
		}
		*windows = append(*windows, CompositionBridgeWindow{row[2], row[3], row[4], row[5], row[6]})
	}
	return nil
}

func validateRecordedBridgeObservations(r CompositionRuntime) error {
	if r.BridgeObservation == nil {
		for _, t := range r.Traces {
			if t.Bridge != nil {
				return fmt.Errorf("recorded bridge trace has no summary")
			}
		}
		return nil
	}
	if r.FieldObservation == nil || r.BridgeObservation.Schema != "gooo/native-field-local-bridge-summary/v1" || len(r.BridgeObservation.SHA256) != 71 || !strings.HasPrefix(r.BridgeObservation.SHA256, "sha256:") {
		return fmt.Errorf("recorded bridge has no field observation or a different schema")
	}
	summary, err := validateBridgeObservations(r.Traces, r)
	if err != nil {
		return err
	}
	summary.SHA256, summary.Scope = r.BridgeObservation.SHA256, r.BridgeObservation.Scope
	if !reflect.DeepEqual(summary, r.BridgeObservation) {
		return fmt.Errorf("recorded bridge summary differs from windows, clocks or source metadata")
	}
	return nil
}
