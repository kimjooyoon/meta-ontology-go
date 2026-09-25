package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestCITerminalJobSnapshotRejectsInProgress(t *testing.T) {
	jobs := make([]jobInput, len(proofJobs))
	head := strings.Repeat("a", 40)
	for index, name := range proofJobs {
		jobs[index] = jobInput{ID: int64(index + 1), Name: name, Status: "completed", Conclusion: "success", HeadSHA: head}
	}
	jobs[len(jobs)-1].Status = "in_progress"
	data, err := json.Marshal(jobs)
	if err != nil {
		t.Fatal(err)
	}
	filename := t.TempDir() + "/jobs.json"
	if err := os.WriteFile(filename, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readJobs(filename); err == nil {
		t.Fatal("in-progress canonical job was accepted")
	}
}
