package main

import (
	"encoding/json"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"io"
	"strconv"
)

func packageOriginQuery(flags map[string]string) (bodyexecution.NativeOriginQuery, error) {
	query := bodyexecution.NativeOriginQuery{Kind: "elements", Limit: 32, RootActivityID: flags["--origin-root-activity-id"], ActivityID: flags["--origin-activity-id"], FieldID: flags["--origin-field-id"]}
	if flags["--origin-receipt"] == "" {
		if hasPackageOriginQueryOptions(flags) {
			return query, fmt.Errorf("origin query options require --origin-receipt")
		}
		return query, nil
	}
	if query.RootActivityID != "" || query.ActivityID != "" || query.FieldID != "" {
		query.Kind = "field-input-locations"
	}
	if flags["--origin-kind"] != "" {
		query.Kind = flags["--origin-kind"]
	}
	shape := flags["--origin-shape"]
	var err error
	query, err = configurePackageOriginShape(query, shape, flags["--origin-kind"] == "")
	if err != nil {
		return query, err
	}
	for _, option := range []struct {
		key    string
		target *int
	}{{"--origin-offset", &query.Offset}, {"--origin-limit", &query.Limit}} {
		if value := flags[option.key]; value != "" {
			n, err := strconv.Atoi(value)
			if err != nil {
				return query, fmt.Errorf("%s requires an integer", option.key)
			}
			*option.target = n
		}
	}
	return query, validatePackageOriginShape(query, shape)
}

func configurePackageOriginShape(query bodyexecution.NativeOriginQuery, shape string, defaultKind bool) (bodyexecution.NativeOriginQuery, error) {
	switch shape {
	case "", "rows", "local-fill-columns", "field-lineage-columns", "field-local-columns",
		"field-local-columns-v2", "field-parent-columns", "local-version-columns", "local-use-columns":
	default:
		return query, fmt.Errorf("origin shape must be rows, local-fill-columns, field-lineage-columns, field-local-columns, field-local-columns-v2, field-parent-columns, local-version-columns or local-use-columns")
	}
	if shape == "local-use-columns" {
		if defaultKind {
			query.Kind = "local-uses"
		}
		query.Limit = 1
	}
	if shape == "local-version-columns" && defaultKind {
		query.Kind = "local-writes"
	}
	if shape == "local-fill-columns" {
		if defaultKind {
			query.Kind = "local-fill-fragments"
		}
		if query.Kind != "local-fill-fragments" {
			return query, fmt.Errorf("local-fill-columns requires local-fill-fragments")
		}
	}
	if shape == "field-lineage-columns" || shape == "field-local-columns" || shape == "field-local-columns-v2" || shape == "field-parent-columns" {
		if defaultKind {
			query.Kind = "field-uses"
		}
		query.Limit = 1
	}
	return query, nil
}

func validatePackageOriginShape(query bodyexecution.NativeOriginQuery, shape string) error {
	if (shape == "field-local-columns" || shape == "field-local-columns-v2") && query.Kind != "field-uses" {
		return fmt.Errorf("%s requires field-uses", shape)
	}
	if (shape == "field-lineage-columns" || shape == "field-parent-columns") && (query.Kind != "field-uses" || query.Limit != 1) {
		return fmt.Errorf("%s requires field-uses and limit1", shape)
	}
	if shape == "local-version-columns" && query.Kind != "local-writes" {
		return fmt.Errorf("local-version-columns requires local-writes")
	}
	if shape == "local-use-columns" && query.Kind != "local-uses" {
		return fmt.Errorf("local-use-columns requires local-uses")
	}
	return bodyexecution.ValidateNativeOriginQuery(query)
}

func packageOriginInputs(raw []byte, target packageruntime.EntrySpec, query bodyexecution.NativeOriginQuery) ([]byte, *bodyexecution.NativeOriginInput, error) {
	return packageOriginInputsWithShape(raw, target, query, "rows")
}

func packageOriginInputsWithShape(raw []byte, target packageruntime.EntrySpec, query bodyexecution.NativeOriginQuery, shape string) ([]byte, *bodyexecution.NativeOriginInput, error) {
	runtime, err := decodePackageOriginRuntime(raw)
	if err != nil {
		return nil, nil, fmt.Errorf("decode origin receipt: %w", err)
	}
	var rows []json.RawMessage
	var origin *bodyexecution.NativeOriginInput
	if shape == "local-use-columns" {
		var columns bodyexecution.LocalUseColumns
		columns, origin, err = bodyexecution.SelectRecordedLocalUses(runtime, query, workspaceDigest(raw))
		if err == nil {
			var row []byte
			row, err = json.Marshal(columns)
			rows = []json.RawMessage{row}
		}
	} else if shape == "local-version-columns" {
		var columns bodyexecution.LocalVersionColumns
		columns, origin, err = bodyexecution.SelectRecordedLocalVersions(runtime, query)
		if err == nil {
			var row []byte
			row, err = json.Marshal(columns)
			rows = []json.RawMessage{row}
		}
	} else if shape == "field-local-columns" || shape == "field-local-columns-v2" {
		rows, origin, err = packageBridgeOriginRows(runtime, query, shape == "field-local-columns-v2")
	} else if shape == "field-parent-columns" {
		var columns bodyexecution.FieldParentColumns
		columns, origin, err = bodyexecution.SelectRecordedFieldParent(runtime, query, workspaceDigest(raw))
		if err == nil {
			var row []byte
			row, err = json.Marshal(columns)
			rows = []json.RawMessage{row}
		}
	} else if shape == "field-lineage-columns" {
		var columns bodyexecution.FieldLineageColumns
		columns, origin, err = bodyexecution.SelectRecordedFieldLineage(runtime, query)
		if err == nil {
			var row []byte
			row, err = json.Marshal(columns)
			rows = []json.RawMessage{row}
		}
	} else {
		selectPage := bodyexecution.SelectRecordedNativeOrigins
		if shape == "local-fill-columns" {
			selectPage = bodyexecution.SelectRecordedNativeOriginPage
		}
		rows, origin, err = selectPage(runtime, query)
		if err == nil {
			rows, err = shapePackageOriginRows(rows, origin, shape)
		}
	}
	if err != nil {
		return nil, origin, err
	}
	return encodePackageOriginInputs(raw, target, query.Limit, rows, origin)
}

func encodePackageOriginInputs(raw []byte, target packageruntime.EntrySpec, limit int, rows []json.RawMessage, origin *bodyexecution.NativeOriginInput) ([]byte, *bodyexecution.NativeOriginInput, error) {
	origin.ReceiptSHA256 = workspaceDigest(raw)
	input := struct {
		Schema string                       `json:"schema"`
		Inputs []map[string]json.RawMessage `json:"inputs"`
	}{Schema: bodyexecution.CompositionInputsSchema}
	for _, row := range rows {
		input.Inputs = append(input.Inputs, map[string]json.RawMessage{target.PackagePath + ":" + target.Activity: row})
	}
	encoded, err := json.Marshal(input)
	if err == nil && len(encoded) > bodyexecution.CompositionOriginPageMaxBytes {
		return nil, origin, fmt.Errorf("selected origin inputs exceed %d bytes; lower --origin-limit (currently %d)", bodyexecution.CompositionOriginPageMaxBytes, limit)
	}
	return encoded, origin, err
}

func writePackageOriginInput(output io.Writer, origin *bodyexecution.NativeOriginInput) {
	if origin == nil {
		return
	}
	fmt.Fprintf(output, "recorded origins: kind=%s offset=%d returned=%d total=%d receipt=%s\n", origin.Query.Kind, origin.Query.Offset, origin.ReturnedRows, origin.TotalRows, origin.ReceiptSHA256)
	if origin.SourceRows != nil {
		fmt.Fprintf(output, "origin subjects: retained_rows=%d selected_rows=%d; paging uses the same subject IDs\n", *origin.SourceRows, origin.TotalRows)
	}
	if origin.InputShape != "" {
		fmt.Fprintf(output, "consumer inputs: shape=%s rows=%d\n", origin.InputShape, origin.ConsumerRows)
	}
	if origin.NextOffset != nil {
		fmt.Fprintf(output, "next page: --origin-offset %d --origin-limit %d\n", *origin.NextOffset, origin.Query.Limit)
	}
	if summary := origin.Observation; summary != nil {
		fmt.Fprintf(output, "original observation: %s unresolved=%d read_drops=%d registration_drops=%d construction_drops=%d interrupted=%d\n", summary.Status, summary.UnresolvedReads, summary.DroppedReads, summary.RegistrationDropped, summary.ConstructionDropped, summary.InterruptedConstructions)
	}
	if summary := origin.LocalObservation; summary != nil {
		fmt.Fprintf(output, "original local observation: %s writes=%d uses=%d unresolved=%d write_drops=%d use_drops=%d frame_drops=%d interrupted=%d\n", summary.Status, summary.RecordedWrites, summary.RecordedUses, summary.Unresolved, summary.WritesDropped, summary.UsesDropped, summary.FramesDropped, summary.Interrupted)
	}

	if summary := origin.FieldObservation; summary != nil {
		fmt.Fprintf(output, "original field history: %s writes=%d uses=%d parents=%d unknown_parents=%d shared_rhs=%d write_drops=%d use_drops=%d frame_drops=%d\n", summary.Status, summary.RecordedWrites, summary.RecordedUses, summary.ParentLinks, summary.UnknownParents, summary.SharedRHSWrites, summary.WritesDropped, summary.UsesDropped, summary.FramesDropped)
	}
	if summary := origin.BridgeObservation; summary != nil {
		fmt.Fprintf(output, "original field/local bridge: %s field_windows=%d local_windows=%d cross_reads=%d completed=%d interrupted=%d\n", summary.Status, summary.FieldWindows, summary.LocalWindows, summary.CrossReads, summary.Completions, summary.Interrupted)
	}
}
