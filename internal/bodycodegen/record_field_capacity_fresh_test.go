package bodycodegen

import (
	"fmt"
	"go/token"
	"go/types"
	"testing"
)

func TestRecordValueSlots10471(t *testing.T) {
	for _, n := range []int{32, 33} {
		fields := make([]*types.Var, n)
		for i := range fields {
			fields[i] = types.NewField(token.NoPos, nil, fmt.Sprintf("F%d", i), types.Typ[types.Int64], false)
		}
		named := types.NewNamed(types.NewTypeName(token.NoPos, nil, "Wide10471", nil), types.NewStruct(fields, nil), nil)
		value, err := zeroRecordBodyValue(named)
		if n == 33 {
			if err == nil {
				t.Fatal("33 interpreted fields accepted")
			}
			continue
		}
		if err != nil || value.Count != 32 || len(value.Values) != RecordFieldLimit {
			t.Fatal("interpreter slots", err)
		}
		value.Values[31].Integer = 10471
		if got := recordBodyScalarValue(value.Values[31]); got != int64(10471) {
			t.Fatal("last scalar slot", got)
		}
		var flow flowValue
		flow.fields[31] = 511
		if len(flow.fields) != RecordFieldLimit || flow.fields[31] != 511 {
			t.Fatal("symbolic field slots")
		}
	}
	t.Log("32 typed interpreter and symbolic slots;33 rejected;bounded arrays")
}
