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
		if flags["--origin-kind"] != "" || flags["--origin-offset"] != "" || flags["--origin-limit"] != "" {
			return query, fmt.Errorf("origin query options require --origin-receipt")
		}
		return query, nil
	}
	if flags["--origin-kind"] != "" {
		query.Kind = flags["--origin-kind"]
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
	return query, bodyexecution.ValidateNativeOriginQuery(query)
}

func packageOriginInputs(raw []byte, target packageruntime.EntrySpec, query bodyexecution.NativeOriginQuery) ([]byte, *bodyexecution.NativeOriginInput, error) {
	runtime, err := decodePackageOriginRuntime(raw)
	if err != nil {
		return nil, nil, fmt.Errorf("decode origin receipt: %w", err)
	}
	rows, origin, err := bodyexecution.SelectRecordedNativeOrigins(runtime, query)
	if err != nil {
		return nil, nil, err
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
	if origin.NextOffset != nil {
		fmt.Fprintf(output, "next page: --origin-offset %d --origin-limit %d\n", *origin.NextOffset, origin.Query.Limit)
	}
	if summary := origin.Observation; summary != nil {
		fmt.Fprintf(output, "original observation: %s unresolved=%d read_drops=%d registration_drops=%d construction_drops=%d interrupted=%d\n", summary.Status, summary.UnresolvedReads, summary.DroppedReads, summary.RegistrationDropped, summary.ConstructionDropped, summary.InterruptedConstructions)
	}
	if summary := origin.LocalObservation; summary != nil {
		fmt.Fprintf(output, "original local observation: %s writes=%d uses=%d unresolved=%d write_drops=%d use_drops=%d frame_drops=%d interrupted=%d\n", summary.Status, summary.RecordedWrites, summary.RecordedUses, summary.Unresolved, summary.WritesDropped, summary.UsesDropped, summary.FramesDropped, summary.Interrupted)
	}

}
