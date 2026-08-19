package selectiveci

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestStrictJSONCodecsRoundTripAndRejectExtraFields(t *testing.T) {
	input := completeFixture().input
	data, err := EncodeInput(input)
	if err != nil {
		t.Fatalf("encode input: %v", err)
	}
	decoded, err := DecodeInput(data)
	if err != nil {
		t.Fatalf("decode input: %v", err)
	}
	if got := Evaluate(decoded); got.Status != Verified {
		t.Fatalf("decoded input receipt = %#v", got)
	}
	receipt := Evaluate(input)
	receiptData, err := EncodeReceipt(receipt)
	if err != nil {
		t.Fatalf("encode receipt: %v", err)
	}
	decodedReceipt, err := DecodeReceipt(receiptData)
	if err != nil || decodedReceipt.Digest != receipt.Digest {
		t.Fatalf("receipt round trip = %#v, err=%v", decodedReceipt, err)
	}
	var wire map[string]any
	if err := json.Unmarshal(data, &wire); err != nil {
		t.Fatal(err)
	}
	wire["unexpected"] = true
	withExtra, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeInput(withExtra); err == nil {
		t.Fatal("unknown input field was accepted")
	}
	duplicate := bytes.Replace(data, []byte(`"schema"`), []byte(`"schema":"`+SchemaVersion+`","schema"`), 1)
	if _, err := DecodeInput(duplicate); err == nil {
		t.Fatal("duplicate input field was accepted")
	}
}
func TestReceiptJSONHasCanonicalOrdering(t *testing.T) {
	receipt := Evaluate(completeFixture().input)
	first, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	second, err := json.Marshal(receipt)
	if err != nil || !reflect.DeepEqual(first, second) {
		t.Fatalf("receipt JSON is not deterministic: %s / %s", first, second)
	}
}
