package eligibilityregistry

import (
	"bytes"
	"testing"
)

func requireFrameBytes(t *testing.T, name string, got, want []byte) {
	t.Helper()
	requireFrame(t, name, bytes.Equal(got, want))
}
func requireFrame(t *testing.T, name string, condition bool) {
	t.Helper()
	if !condition {
		t.Fatal(name)
	}
}
func TestFrameFieldNilAndEmpty(t *testing.T) {
	requireFrame(t, "nil value", encodeField(frameField{name: "x", tag: frameTagString}) == nil)
	empty := encodeField(frameField{name: "x", tag: frameTagString, value: []byte{}})
	want := []byte{0, 0, 0, 0, 0, 0, 0, 1, 'x', 0x01, 0, 0, 0, 0, 0, 0, 0, 0}
	requireFrameBytes(t, "empty value", empty, want)
	requireFrame(t, "invalid tag", encodeField(frameField{tag: 0x04, value: []byte{}}) == nil)
	nilFields, emptyFields := encodeFrame("d", nil), encodeFrame("d", []frameField{})
	requireFrame(t, "nil fields", nilFields != nil && len(nilFields) == 17)
	requireFrameBytes(t, "empty fields", emptyFields, nilFields)
	requireFrame(t, "nil frame field", encodeFrame("d", []frameField{{tag: frameTagString}}) == nil)
}
func TestEncodeFieldAndFrameVectors(t *testing.T) {
	field := encodeField(frameField{name: "id", tag: frameTagStableID, value: []byte("x")})
	wantField := []byte{0, 0, 0, 0, 0, 0, 0, 2, 'i', 'd', 0x02, 0, 0, 0, 0, 0, 0, 0, 1, 'x'}
	requireFrameBytes(t, "field", field, wantField)
	emptyName := encodeField(frameField{tag: frameTagEnum, value: []byte("PASS")})
	wantEmptyName := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0x05, 0, 0, 0, 0, 0, 0, 0, 4, 'P', 'A', 'S', 'S'}
	requireFrameBytes(t, "empty name", emptyName, wantEmptyName)
	frame := encodeFrame("d", []frameField{
		{name: "a", tag: frameTagString, value: []byte("1")},
		{name: "b", tag: frameTagDigest, value: []byte("2")},
	})
	wantFrame := []byte{
		0, 0, 0, 0, 0, 0, 0, 1, 'd', 0, 0, 0, 0, 0, 0, 0, 2,
		0, 0, 0, 0, 0, 0, 0, 1, 'a', 0x01, 0, 0, 0, 0, 0, 0, 0, 1, '1',
		0, 0, 0, 0, 0, 0, 0, 1, 'b', 0x03, 0, 0, 0, 0, 0, 0, 0, 1, '2',
	}
	requireFrameBytes(t, "ordered frame", frame, wantFrame)
}
