package bodycodegen

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
)

func TestRecordScopeChecksRealizedCandidatesBeforeEvaluation(t *testing.T) {
	model := bidir.Model{}
	for i := range 17 {
		model.Nodes = append(model.Nodes, recordShapeFixture(fmt.Sprintf("R%d", i), 1))
	}
	records, err := DeclaredRecordTypesFromModel(model)
	if err != nil {
		t.Fatal(err)
	}
	var source strings.Builder
	source.WriteString("package candidate\n" + RecordDeclarations(records, false))
	source.WriteString("func Evaluate(input R0) R0 {\nvar total int64 = input.value0\n")
	for i := 1; i < 17; i++ {
		fmt.Fprintf(&source, "r%d := R%d{value0: input.value0}\ntotal += r%d.value0\n", i, i, i)
	}
	source.WriteString("return R0{value0: total}\n}\n")
	raw := []byte(source.String())
	if _, err := evaluateRecordAssembly(context.Background(), raw, "Evaluate", records, nil); err == nil || !strings.Contains(err.Error(), "typed body") {
		t.Fatal("candidate scorer accepted widened record use", err)
	}
	if _, err := newCriteriaTupleEvaluator(context.Background(), raw, "Evaluate", records, nil); err == nil || !strings.Contains(err.Error(), "typed body") {
		t.Fatal("criteria scorer accepted widened record use", err)
	}
}
