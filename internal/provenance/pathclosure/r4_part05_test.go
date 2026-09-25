package pathclosure_test

import (
	"bytes"
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance/pathclosure"
	"reflect"
	"testing"
)

func TestStrictR4JSONCodecCanonicalReplay(t *testing.T) {
	fixture := completeR4Fixture()
	data, err := pathclosure.EncodeR4Input(fixture.input)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := pathclosure.DecodeR4Input(data)
	if err != nil || pathclosure.EvaluateR4(decoded).Status != pathclosure.PASS {
		t.Fatalf("round trip = %#v, err=%v", decoded, err)
	}
	if replay, err := pathclosure.EncodeR4Input(decoded); err != nil || !bytes.Equal(data, replay) {
		t.Fatalf("canonical replay changed: %s / %s", data, replay)
	}

	for _, test := range []struct {
		name   string
		mutate func([]byte) []byte
	}{
		{"unknown field", func(value []byte) []byte {
			return bytes.Replace(value, []byte(`"schema"`), []byte(`"unknown":"x","schema"`), 1)
		}},
		{"removed expected label field", func(value []byte) []byte {
			return bytes.Replace(value, []byte(`"record_bytes":[`), []byte(`"expected_labels":[],"record_bytes":[`), 1)
		}},
		{"duplicate key", func(value []byte) []byte {
			return bytes.Replace(value, []byte(`"schema":"`+pathclosure.R4SchemaVersion+`"`), []byte(`"schema":"`+pathclosure.R4SchemaVersion+`","schema":"`+pathclosure.R4SchemaVersion+`"`), 1)
		}},
		{"trailing value", func(value []byte) []byte { return append(append([]byte(nil), value...), []byte(` {}`)...) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := pathclosure.DecodeR4Input(test.mutate(data)); err == nil {
				t.Fatal("malformed strict JSON was accepted")
			}
		})
	}
	var wire map[string]any
	if err := json.Unmarshal(data, &wire); err != nil {
		t.Fatal(err)
	}
	if _, ok := wire["records"]; !ok {
		t.Fatal("records missing from canonical input")
	}
}
func TestEvaluateR4PermutationReplayUsesStableIDsNotInsertionOrder(t *testing.T) {
	fixture := completeR4Fixture()
	permuted := cloneR4Input(fixture.input)
	permuted.Records[0], permuted.Records[1] = permuted.Records[1], permuted.Records[0]
	permuted.Receipts[0], permuted.Receipts[1] = permuted.Receipts[1], permuted.Receipts[0]
	left, right := pathclosure.EvaluateR4(fixture.input), pathclosure.EvaluateR4(permuted)
	if !reflect.DeepEqual(left, right) {
		t.Fatalf("permutation changed R4 result:\nleft=%#v\nright=%#v", left, right)
	}
}
