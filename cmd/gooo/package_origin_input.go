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
	query := bodyexecution.NativeOriginQuery{Kind: "elements", Limit: 32}
	if flags["--origin-receipt"] == "" {
		if flags["--origin-kind"] != "" || flags["--origin-offset"] != "" || flags["--origin-limit"] != "" || flags["--origin-shape"] != "" {
			return query, fmt.Errorf("origin query options require --origin-receipt")
		}
		return query, nil
	}
	if flags["--origin-kind"] != "" {
		query.Kind = flags["--origin-kind"]
	}
	shape := flags["--origin-shape"]
	if shape != "" && shape != "rows" && shape != "local-fill-columns" && shape != "field-lineage-columns" {
		return query, fmt.Errorf("origin shape must be rows, local-fill-columns or field-lineage-columns")
	}
	if shape == "local-fill-columns" {
		if flags["--origin-kind"] == "" {
			query.Kind = "local-fill-fragments"
		}
		if query.Kind != "local-fill-fragments" {
			return query, fmt.Errorf("local-fill-columns requires local-fill-fragments")
		}
	}
	if shape == "field-lineage-columns" {
		if flags["--origin-kind"] == "" {
			query.Kind = "field-uses"
		}
		query.Limit = 1
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
	if shape == "field-lineage-columns" && (query.Kind != "field-uses" || query.Limit != 1) {
		return query, fmt.Errorf("field-lineage-columns requires field-uses and limit1")
	}
	return query, bodyexecution.ValidateNativeOriginQuery(query)
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
	if shape == "field-lineage-columns" {
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
	origin.ReceiptSHA256 = workspaceDigest(raw)
	input := struct {
		Schema string                       `json:"schema"`
		Inputs []map[string]json.RawMessage `json:"inputs"`
	}{Schema: bodyexecution.CompositionInputsSchema}
	for _, row := range rows {
		input.Inputs = append(input.Inputs, map[string]json.RawMessage{target.PackagePath + ":" + target.Activity: row})
	}
	encoded, err := json.Marshal(input)
	if err == nil && len(encoded) > 32<<10 {
		return nil, origin, fmt.Errorf("selected origin inputs exceed 32768 bytes; lower --origin-limit (currently %d)", query.Limit)
	}
	return encoded, origin, err
}

func writePackageOriginInput(output io.Writer, origin *bodyexecution.NativeOriginInput) {
	if origin == nil {
		return
	}
	fmt.Fprintf(output, "recorded origins: kind=%s offset=%d returned=%d total=%d receipt=%s\n", origin.Query.Kind, origin.Query.Offset, origin.ReturnedRows, origin.TotalRows, origin.ReceiptSHA256)
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

}
