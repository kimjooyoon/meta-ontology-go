package main

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func packageBridgeOriginRows(r bodyexecution.CompositionRuntime, q bodyexecution.NativeOriginQuery, indexed bool) ([]json.RawMessage, *bodyexecution.NativeOriginInput, error) {
	if indexed {
		columns, report, err := bodyexecution.SelectRecordedIndexedBridgeLineagePage(r, q)
		if err != nil {
			return nil, report, err
		}
		rows, err := marshalPackageColumnRows(columns)
		return rows, report, err
	}
	columns, report, err := bodyexecution.SelectRecordedBridgeLineagePage(r, q)
	if err != nil {
		return nil, report, err
	}
	rows, err := marshalPackageColumnRows(columns)
	return rows, report, err
}

func marshalPackageColumnRows[T any](columns []T) ([]json.RawMessage, error) {
	rows := make([]json.RawMessage, len(columns))
	for i, column := range columns {
		raw, err := json.Marshal(column)
		if err != nil {
			return nil, err
		}
		rows[i] = raw
	}
	return rows, nil
}
