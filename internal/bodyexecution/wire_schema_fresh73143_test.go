package bodyexecution

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"
)

type wireRow73143 struct {
	Origin string          `json:"origin"`
	Number json.Number     `json:"number"`
	Bytes  []byte          `json:"bytes"`
	Detail json.RawMessage `json:"detail"`
}

type wireReceipt73143 struct {
	Rows   []wireRow73143           `json:"rows"`
	Lookup map[string]*wireRow73143 `json:"lookup"`
	Head   *wireRow73143            `json:"head"`
	Fixed  [1]wireRow73143          `json:"fixed"`
	Free   any                      `json:"free"`
}

func TestReceiptWireShapes73143(t *testing.T) {
	row := `{"origin":"기원/73143","number":9007199254740993,"bytes":"AQI=","detail": {"arbitrary":73143}}`
	raw := []byte(`{"rows":[` + row + `],"lookup":{"기원":` + row + `},"head":` + row + `,"fixed":[` + row + `],"free":{"custom":[73143]}}`)
	var receipt wireReceipt73143
	if err := DecodeExecutionReceipt(raw, &receipt); err != nil {
		t.Fatal(err)
	}
	if len(receipt.Rows) != 1 || receipt.Rows[0].Number.String() != "9007199254740993" ||
		receipt.Lookup["기원"].Origin != "기원/73143" || receipt.Head.Origin != receipt.Fixed[0].Origin ||
		!bytes.Equal(receipt.Head.Bytes, []byte{1, 2}) ||
		!bytes.Equal(receipt.Head.Detail, []byte(`{"arbitrary":73143}`)) {
		t.Fatal("source fields, numbers or opaque wire bytes changed", receipt)
	}
	var absent wireReceipt73143
	if err := DecodeExecutionReceipt([]byte(`{"rows":null,"lookup":null,"head":null,"free":null}`), &absent); err != nil {
		t.Fatal("existing null contract changed", err)
	}
}

func TestReceiptWireCounterexamples73143(t *testing.T) {
	for _, raw := range wireCounterexamples73143() {
		var receipt wireReceipt73143
		if err := DecodeExecutionReceipt([]byte(raw), &receipt); err == nil {
			t.Fatal("wire ambiguity or invalid shape accepted", raw)
		}
	}
	var receipt wireReceipt73143
	if err := DecodeExecutionReceipt([]byte{'{', '"', 'f', 'r', 'e', 'e', '"', ':', '"', 0xff, '"', '}'}, &receipt); err == nil {
		t.Fatal("invalid UTF-8 accepted")
	}
	if err := DecodeExecutionReceipt([]byte(strings.Repeat(" ", (32<<20)+1)), &receipt); err == nil {
		t.Fatal("32 MiB input bound changed")
	}
	if err := DecodeExecutionReceipt([]byte(`{}`), (*wireReceipt73143)(nil)); err == nil {
		t.Fatal("nil target accepted")
	}
}

func wireCounterexamples73143() []string {
	return []string{
		`{"Rows":[]}`, `{"rows":[{"Origin":"x"}]}`, `{"lookup":{"x":{"Origin":"x"}}}`,
		`{"head":{"Origin":"x"}}`, `{"fixed":[{"Origin":"x"}]}`,
		`{"rows":[],"rows":[]}`, `{"rows":[],"\u0072ows":[]}`,
		`{"rows":[{"detail":{"x":1,"x":2}}]}`, `{"free":{"x":1,"x":2}}`,
		`{"rows":[{"bytes":[1,2]}]}`, `{"rows":[{"bytes":"!"}]}`,
		`{"rows":{}}`, `{"head":73143}`, `{"missing":73143}`, `{} {}`,
		`{"free":` + strings.Repeat("[", 65) + "73143" + strings.Repeat("]", 65) + `}`,
	}
}

func TestReceiptWireSourceTraversal73143(t *testing.T) {
	row := `{"origin":"urn:gooo:source:73143","number":73143,"bytes":"AQI=","detail":{"note":"` + strings.Repeat("a", 1536) + `"}}`
	raw := []byte(`{"rows":[` + strings.Repeat(row+",", 4095) + row + `]}`)
	scanner := json.NewDecoder(bytes.NewReader(raw))
	scanner.UseNumber()
	schema := wireSchema{fields: map[reflect.Type]map[string]reflect.Type{}}
	if err := schema.scanValue(scanner, 0, reflect.TypeFor[wireReceipt73143]()); err != nil {
		t.Fatal(err)
	}
	if _, err := scanner.Token(); err != io.EOF || len(schema.fields) != 2 {
		t.Fatal("shape table depends on row count", len(schema.fields), err)
	}
	start := time.Now()
	var receipt wireReceipt73143
	if err := DecodeExecutionReceipt(raw, &receipt); err != nil || len(receipt.Rows) != 4096 {
		t.Fatal("new source-record population", len(receipt.Rows), err)
	}
	for _, value := range receipt.Rows {
		if value.Origin != "urn:gooo:source:73143" || value.Number.String() != "73143" || len(value.Detail) != 1547 {
			t.Fatal("source row changed", value)
		}
	}
	t.Logf("new synthetic receipt bytes=%d rows=4096 shape_types=2 decode_wall=%s;one invocation,no runtime/model calls", len(raw), time.Since(start))
}
