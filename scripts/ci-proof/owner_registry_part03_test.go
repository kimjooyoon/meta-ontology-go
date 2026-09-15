package main

import (
	"encoding/json"
	"testing"
)

func failureOwnerRegistryDocument(t *testing.T, mutate func(*failureOwnerRegistry)) []byte {
	t.Helper()
	data, err := readFailureFile(failureOwnerRegistryPath)
	if err != nil {
		t.Fatal(err)
	}
	var registry failureOwnerRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		t.Fatal(err)
	}
	if mutate != nil {
		mutate(&registry)
	}
	data, err = json.Marshal(registry)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
