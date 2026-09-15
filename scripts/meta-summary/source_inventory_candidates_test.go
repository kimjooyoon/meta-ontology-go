package main

import "testing"

func TestDecodeSelectedSubjectsRequiresGenerationV7Shape(t *testing.T) {
	valid := []byte(`{"schema_version":"gooo/self-improvement-generation/v7","selected":[{"meta_operation":"extract-function","metric_id":"gooo.metric.source.function-lines.v1","subject":"fixture.go:1:Fixture"}]}`)
	selected, err := decodeSelectedSubjects(valid)
	if err != nil || len(selected) != 1 || selected[0].MetaOperation != "extract-function" {
		t.Fatalf("valid v7 selected shape was rejected: selected=%#v err=%v", selected, err)
	}

	missingOperation := []byte(`{"schema_version":"gooo/self-improvement-generation/v6","selected":[{"operation":"extract-function","metric_id":"gooo.metric.source.function-lines.v1","subject":"fixture.go:1:Fixture"}]}`)
	if _, err := decodeSelectedSubjects(missingOperation); err == nil {
		t.Fatal("legacy operation field was accepted without meta_operation")
	}

	wrongSchema := []byte(`{"schema_version":"gooo/self-improvement-generation/v5","selected":[{"meta_operation":"extract-function","metric_id":"gooo.metric.source.function-lines.v1","subject":"fixture.go:1:Fixture"}]}`)
	if _, err := decodeSelectedSubjects(wrongSchema); err == nil {
		t.Fatal("wrong self-improvement generation schema was accepted")
	}
}
