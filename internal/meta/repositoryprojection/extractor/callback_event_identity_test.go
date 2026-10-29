package extractor

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestCallbackTestEventKeepsDeclaredNominalIdentity(t *testing.T) {
	kind := reflect.TypeFor[CallbackPackageTestEvent]()
	const owner = "github.com/kimjooyoon/meta-ontology-go/internal/meta/repositoryprojection/extractor"
	if kind.Name() != "CallbackPackageTestEvent" || kind.PkgPath() != owner {
		t.Fatal("transport refactor changed the existing observation type identity", kind)
	}
	raw, err := json.Marshal(CallbackPackageTestEvent{Name: "TestSubject", Action: "fail"})
	if err != nil || string(raw) != `{"name":"TestSubject","action":"fail"}` {
		t.Fatal("transport refactor changed the existing event wire", string(raw), err)
	}
}
