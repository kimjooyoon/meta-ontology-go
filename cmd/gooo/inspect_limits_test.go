package main

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestMarshalGraphDumpRejectsOversizeOutput(t *testing.T) {
	ir := semantic.NewIR("billing", semantic.Namespace("billing"))
	for index := 0; index < 15000; index++ {
		id := semantic.MustIdentity(fmt.Sprintf("billing://entity/item-%05d", index))
		if err := ir.AddNode(semantic.Node{ID: id, Kind: semantic.Entity, Namespace: "billing", Name: fmt.Sprintf("Item%05d", index)}); err != nil {
			t.Fatal(err)
		}
	}
	_, err := marshalGraphDump(nil, ir)
	if !errors.Is(err, errGraphDumpLimit) {
		t.Fatalf("oversize graph error = %v, want %v", err, errGraphDumpLimit)
	}
}

func TestWriteInspectOutputHonorsDeadlineAndDeadlineWriter(t *testing.T) {
	var output deadlineBuffer
	deadline := time.Now().Add(time.Second)
	if err := writeInspectOutput(&output, []byte("ok"), deadline); err != nil {
		t.Fatal(err)
	}
	if len(output.deadlines) != 2 || !output.deadlines[0].Equal(deadline) || !output.deadlines[1].IsZero() {
		t.Fatalf("writer deadlines = %#v", output.deadlines)
	}
	var expired bytes.Buffer
	if err := writeInspectOutput(&expired, []byte("blocked"), time.Now().Add(-time.Second)); !errors.Is(err, errCommandDeadline) || expired.Len() != 0 {
		t.Fatalf("expired output = err %v, bytes %d", err, expired.Len())
	}
}

type deadlineBuffer struct {
	bytes.Buffer
	deadlines []time.Time
}

func (b *deadlineBuffer) SetWriteDeadline(deadline time.Time) error {
	b.deadlines = append(b.deadlines, deadline)
	return nil
}
