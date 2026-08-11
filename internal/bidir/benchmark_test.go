package bidir

import (
	"fmt"
	"testing"
)

func BenchmarkMeasureBXFixture(b *testing.B) {
	fixture := billingBXFixture{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := MeasureBXFixture(fixture); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReconcileScaling(b *testing.B) {
	for _, edgeCount := range []int{1, 8, 32} {
		b.Run(fmt.Sprintf("edges=%d", edgeCount), func(b *testing.B) {
			model, err := Get(benchmarkDocument(edgeCount + 1))
			if err != nil {
				b.Fatal(err)
			}
			delta := benchmarkDelta(edgeCount)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := Reconcile(model, delta); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func benchmarkDocument(activityCount int) Document {
	document := Document{Package: "benchmark", Namespace: "benchmark"}
	for index := 0; index < activityCount; index++ {
		id := ID(fmt.Sprintf("benchmark://activity/a-%03d", index))
		document.Declarations = append(document.Declarations, Declaration{
			Kind: ActivityKind,
			ID:   id,
			Name: fmt.Sprintf("Activity%03d", index),
		})
	}
	return document
}

func benchmarkDelta(edgeCount int) FactDelta {
	facts := make(FactSet, edgeCount)
	for index := 0; index < edgeCount; index++ {
		fact := NewSourcedFact(
			DeterministicFact,
			ID(fmt.Sprintf("benchmark://activity/a-%03d", index)),
			PredicateInvokes,
			ID(fmt.Sprintf("benchmark://activity/a-%03d", index+1)),
			SourceSpan{File: "benchmark.go", Start: index + 1, End: index + 2},
		)
		fact.SubjectKind = ActivityKind
		fact.ObjectKind = ActivityKind
		facts[index] = fact
	}
	return FactDelta{Added: facts}
}
