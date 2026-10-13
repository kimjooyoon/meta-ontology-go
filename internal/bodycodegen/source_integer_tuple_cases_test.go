package bodycodegen

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

func TestSourceIntegerTupleCaseTypes6533(t *testing.T) {
	const source = `package tuple
namespace tuple6533
entity Integer id "urn:gooo:type:integer"
activity Sum(Integer,Integer) -> Integer computes "return input0 + input1"
`
	for _, input := range []string{`[1]`, `[1,2,3]`, `[1,null]`, `[1,"2"]`, `[1,2.5]`, `[1,9223372036854775808]`} {
		spec := &assemblyspec.Spec{ValueCases: []assemblyspec.ValueCase{{Inputs: input, Expected: "3"}}}
		if _, _, typed, err := sourceIntegerTupleCases("tuple.gooo", []byte(source), "Sum", spec); !typed || err == nil {
			t.Fatal("invalid tuple accepted", input, err)
		}
	}
	spec := &assemblyspec.Spec{ValueCases: []assemblyspec.ValueCase{{Inputs: "[9007199254740993,6]", Expected: "9007199254740999"}}, ValueHoldoutCases: []assemblyspec.ValueCase{{Inputs: "[-43,17]", Expected: "-26"}}}
	training, holdout, typed, err := sourceIntegerTupleCases("tuple.gooo", []byte(source), "Sum", spec)
	if err != nil || !typed || len(training) != 1 || len(holdout) != 1 || training[0].Inputs[0] != 9007199254740993 || training[0].Expected != 9007199254740999 || holdout[0].Inputs[1] != 17 {
		t.Fatal("exact tuple inputs", training, holdout, err)
	}
	t.Log("6533: six malformed/unsupported integer tuples rejected; int64 above2^53 preserved; holdout remains separate; no body/model execution")
}
