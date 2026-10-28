package bodycodegen

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestFixedProgramCallsHaveIndependentValuesAndContexts(t *testing.T) {
	source := []byte(`package prepared
namespace prepared
entity Integer id "urn:gooo:type:integer"
activity Divide(left Integer, right Integer) -> Integer id "prepared://divide" computes "return left/right"
`)
	compileContext, stopCompile := context.WithCancel(context.Background())
	p, err := NewFixedProgram(compileContext, "prepared.gooo", source, "Divide")
	if err != nil {
		t.Fatal(err)
	}
	sourceHash := digest(source)
	stopCompile()
	copy(source, strings.Repeat("x", len(source)))
	if p.template.evaluator.context != nil || p.template.evaluator.environment != nil || p.template.evaluator.callCount != nil {
		t.Fatal("preparation retained invocation state")
	}
	for _, input := range []string{`[1,0]`, `["bad",1]`, `[1]`, `[1,2,3]`, `not-json`} {
		if _, err := p.Evaluate(context.Background(), []byte(input)); err == nil {
			t.Fatal("invalid call accepted", input)
		}
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := p.Evaluate(cancelled, []byte(`[8,2]`)); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for i := int64(1); i <= 32; i++ {
		group.Go(func() {
			input := fmt.Sprintf("[%d,2]", i*2)
			value, err := p.Evaluate(context.Background(), []byte(input))
			if err != nil || string(value.Value) != fmt.Sprint(i) || value.SourceSHA256 != sourceHash || value.ActivityID != "prepared://divide" || value.ProjectionSHA256 == "" {
				t.Errorf("independent call %d: %s / %v", i, value.Value, err)
			}
		})
	}
	group.Wait()
	if p.template.evaluator.environment != nil || p.template.evaluator.callCount != nil {
		t.Fatal("evaluation changed the shared template")
	}
}

func TestFixedProgramRetainsSourceFieldBounds(t *testing.T) {
	source := []byte(`package prepared
namespace prepared
entity Text id "urn:gooo:type:string"
entity Long id "prepared://long" fields {
 field value id "prepared://long/value" type string required one max_utf8_bytes "1030"
}
activity Echo(value Long) -> Long id "prepared://echo" computes "return value"
`)
	p, err := NewFixedProgram(context.Background(), "bounded.gooo", source, "Echo")
	if err != nil {
		t.Fatal(err)
	}
	for _, size := range []int{1030, 1031, 1030} {
		input := []byte(fmt.Sprintf(`[{"value":%q}]`, strings.Repeat("a", size)))
		value, err := p.Evaluate(context.Background(), input)
		if size == 1031 {
			if err == nil {
				t.Fatal("prepared execution lost the source bound")
			}
			continue
		}
		if err != nil || !strings.Contains(string(value.Value), strings.Repeat("a", size)) {
			t.Fatal(size, err)
		}
	}
	var empty FixedProgram
	for _, p := range []*FixedProgram{nil, &empty} {
		if _, err := p.Evaluate(context.Background(), []byte(`[1]`)); err == nil {
			t.Fatal("unprepared program accepted")
		}
	}
	if _, err := NewFixedProgram(nil, "missing.gooo", source, "Echo"); err == nil {
		t.Fatal("missing context accepted")
	}
}

func TestFixedProgramConcurrentNominalRecords(t *testing.T) {
	for _, activity := range []string{"NewChild", "ThroughAlias"} {
		t.Run(activity, func(t *testing.T) {
			program, err := NewFixedProgram(context.Background(), "nominal.gooo", []byte(nominalOriginSource), activity)
			if err != nil {
				t.Fatal(err)
			}
			var group sync.WaitGroup
			for i := 1; i <= 32; i++ {
				group.Go(func() {
					input := []byte(fmt.Sprintf(`[{"before":{"value":11},"after":{"value":22}},%d]`, i))
					if activity == "ThroughAlias" {
						input = []byte(fmt.Sprintf(`[{"before":{"value":11},"after":{"value":%d}}]`, i))
					}
					value, err := program.Evaluate(context.Background(), input)
					if err != nil || string(value.Value) != fmt.Sprint(i) {
						t.Errorf("nominal invocation %d shared another value: %s / %v", i, value.Value, err)
					}
				})
			}
			group.Wait()
			if program.template.evaluator.environment != nil || program.template.evaluator.callCount != nil {
				t.Fatal("nominal values or helper call counters retained by the program")
			}
		})
	}
}
