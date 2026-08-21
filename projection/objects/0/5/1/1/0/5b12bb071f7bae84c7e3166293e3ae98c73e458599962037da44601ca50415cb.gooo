package main

import (
	"bytes"
	"testing"
)

func TestLoopStagesAreOrdered(t *testing.T) {
	previous := -1
	for frame := range frameCount {
		stage := loopState(frame).stage
		if stage < previous {
			t.Fatalf("stage regressed at frame %d: %d after %d", frame, stage, previous)
		}
		previous = stage
	}
	if previous != stageNextWork {
		t.Fatalf("final stage = %d, want %d", previous, stageNextWork)
	}
}
func TestGeneratedBytesAreDeterministic(t *testing.T) {
	gifA, pngA, err := encodedAssets()
	if err != nil {
		t.Fatal(err)
	}
	gifB, pngB, err := encodedAssets()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gifA, gifB) || !bytes.Equal(pngA, pngB) {
		t.Fatal("repeated generation produced different bytes")
	}
}
