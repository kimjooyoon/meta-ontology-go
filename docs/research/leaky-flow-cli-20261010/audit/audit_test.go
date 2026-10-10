package main

import (
	"strconv"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestOriginalLearnedCompilerRecords(t *testing.T) { audit("../result") }

func mustReject(t *testing.T, check func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("one-unit error beyond 2^53 was accepted")
		}
	}()
	check()
}

func TestSourceAuditRejectsRoundedLargeIntegers(t *testing.T) {
	for _, input := range []int64{-9007199254740995, 9007199254740993} {
		t.Run(strconv.FormatInt(input, 10), func(t *testing.T) {
			root := "../result/max-direct"
			if input < 0 {
				root = "../result/min-direct"
			}
			r := decode[bodycodegen.Result](read(root, "initial.json.gz"))
			found := false
			for i, row := range r.Report.BodyPaths.NativeCases {
				if row.Input == input {
					r.Report.BodyPaths.NativeCases[i].Expected++
					r.Report.BodyPaths.NativeCases[i].Actual++
					found = true
				}
			}
			must(found, "negative test case exists")
			mustReject(t, func() { checkSourceOutputs(root, r.Report.BodyPaths) })
		})
	}
}

func TestNativeAuditRejectsRoundedLargeIntegers(t *testing.T) {
	for _, input := range []int64{-9007199254740995, 9007199254740993} {
		t.Run(strconv.FormatInt(input, 10), func(t *testing.T) {
			root := "../result/max-direct"
			if input < 0 {
				root = "../result/min-direct"
			}
			r := decode[constructionOutput](read(root, "construction.json.gz"))
			suite := decode[bodyexecution.CompositionCases](read(root, "evaluation-cases.json"))
			found := false
			for i, trace := range r.Evaluation.Runtime.Traces {
				if decode[int64](trace.Deliveries[0].Input) == input {
					bad := []byte(strconv.FormatInt(input+1, 10))
					r.Evaluation.Runtime.Traces[i].Deliveries[0].Expected = bad
					r.Evaluation.Runtime.Traces[i].Deliveries[0].Actual = bad
					found = true
				}
			}
			must(found, "negative native case exists")
			mustReject(t, func() { checkNative(r.Evaluation.Runtime, suite) })
		})
	}
}
