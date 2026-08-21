package workfrontier

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"testing"
)

func canonicalResultBytes(t *testing.T, result any) []byte {
	t.Helper()
	for _, name := range []string{"CanonicalBytes", "CanonicalJSON", "Canonical"} {
		value, ok := callResultMethod(t, result, name)
		if !ok {
			continue
		}
		switch value := value.(type) {
		case []byte:
			return append([]byte(nil), value...)
		case string:
			return []byte(value)
		default:
			t.Fatalf("result method %s returned %T, want bytes or string", name, value)
		}
	}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func optionalResultDigest(t *testing.T, result any) (string, bool) {
	t.Helper()
	for _, name := range []string{"CanonicalDigest", "Digest", "StableHash"} {
		value, ok := callResultMethod(t, result, name)
		if !ok {
			continue
		}
		digest, ok := value.(string)
		if !ok {
			t.Fatalf("result method %s returned %T, want string", name, value)
		}
		want := sha256.Sum256(canonicalResultBytes(t, result))
		got := digest
		if len(got) > len("sha256:") && got[:len("sha256:")] == "sha256:" {
			got = got[len("sha256:"):]
		}
		if got != hex.EncodeToString(want[:]) {
			t.Fatalf("result digest = %q, want SHA-256 %q", digest, hex.EncodeToString(want[:]))
		}
		return digest, true
	}
	return "", false
}
func callResultMethod(t *testing.T, result any, name string) (any, bool) {
	t.Helper()
	method := reflect.ValueOf(result).MethodByName(name)
	if !method.IsValid() || method.Type().NumIn() != 0 {
		return nil, false
	}
	values := method.Call(nil)
	if len(values) == 0 || len(values) > 2 {
		t.Fatalf("result method %s has unsupported return count %d", name, len(values))
	}
	if len(values) == 2 {
		if err, ok := values[1].Interface().(error); ok && err != nil {
			t.Fatalf("result method %s error = %v", name, err)
		}
	}
	return values[0].Interface(), true
}
