package bodycodegen

import (
	"go/types"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/textbound"
)

func TestInterpreterTextEscapeAndBoundFresh1905(t *testing.T) {
	for _, tc := range []struct {
		name, raw, want, reason string
		maximum                 uint32
	}{
		{"pair-at", `"\ud83e\udded"`, "🧭", "", 4},
		{"pair-above", `"\ud83e\udded"`, "", "at most 3", 3},
		{"replacement", `"\ufffd"`, "�", "", 3},
		{"nul-at", `"\u0000"`, "\x00", "", 1},
		{"nul-above", `"\u0000"`, "", "at most 0", 0},
		{"unpaired-high", `"\ud800"`, "", "UNPAIRED_HIGH", 1030},
		{"unpaired-low", `"\udfff"`, "", "UNPAIRED_LOW", 1030},
		{"null", `null`, "", "requires a string", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := decodeRecordCaseField([]byte(tc.raw), types.Typ[types.String], nil, textbound.New(tc.maximum))
			if tc.reason != "" {
				if err == nil || !strings.Contains(err.Error(), tc.reason) {
					t.Fatal("missing exact domain refusal", actual, err, tc.reason)
				}
			} else if err != nil || actual != tc.want {
				t.Fatal("decoded source value changed", actual, err)
			}
		})
	}
}

func TestInterpreterOutputByteBoundFresh1905(t *testing.T) {
	record := RecordType{Name: "Box", Fields: []RecordField{{Name: "value", ID: "bound1905://value",
		TypeID: "urn:gooo:type:string", TextBound: textbound.New(1030)}}}
	for _, size := range []int{1030, 1031} {
		value := recordBodyValue{Type: "Box", Count: 1}
		value.Values[0] = recordBodyScalar{Text: strings.Repeat("x", size), Kind: 1, Present: true}
		err := validateRecordOutputBounds(value, record)
		if (err == nil) != (size == 1030) {
			t.Fatal("interpreter result did not follow source byte domain", size, err)
		}
	}
}
