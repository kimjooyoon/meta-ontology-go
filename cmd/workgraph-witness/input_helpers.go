package main

import (
	"encoding/json"
	"os"
	"runtime"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/workgraph"
)

func readOptional(path string) ([]byte, error) {
	if path == "" { return nil, nil }
	return os.ReadFile(path)
}

func optionalDigest(value []byte) string {
	if value == nil { return "" }
	return workgraph.DigestBytes(value)
}

func readPredecessor(path string) (*workgraph.Report, string, error) {
	if path == "" { return nil, "", nil }
	data, err := os.ReadFile(path)
	if err != nil { return nil, "", err }
	var report workgraph.Report
	if err := json.Unmarshal(data, &report); err != nil { return nil, "", err }
	return &report, workgraph.DigestBytes(data), nil
}

func resourceSample(path string, started time.Time, before runtime.MemStats) (workgraph.ResourceSample, error) {
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil { return workgraph.ResourceSample{}, err }
		var sample workgraph.ResourceSample
		if err := json.Unmarshal(data, &sample); err != nil { return sample, err }
		return sample, nil
	}
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	wall := time.Since(started).Nanoseconds()
	if wall <= 0 { wall = 1 }
	return workgraph.ResourceSample{Samples: 1, WallNanoseconds: wall, HeapSysBytes: after.HeapSys, TotalAllocBytes: after.TotalAlloc - before.TotalAlloc}, nil
}
