package main

import (
	"encoding/json"
	"runtime/debug"
	"testing"
)

func TestVersionBuildRetainsInstalledModuleIdentity(t *testing.T) {
	info := &debug.BuildInfo{Main: debug.Module{
		Path:    "github.com/kimjooyoon/meta-ontology-go",
		Version: "v0.6.5-dev.0.20261007174112-25df294860e1",
		Sum:     "h1:Q3T4B9HJ+cbJl9JL9oqeobZmff19pocAet10bCAXVZw=",
	}}
	identity := versionBuildMetadata(info, "go1.27.1")
	raw, err := json.Marshal(identity)
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Module *struct{ Path, Version, Sum string }
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Module == nil {
		t.Fatal("installed module identity is absent from build JSON")
	}
	if payload.Module.Path != info.Main.Path || payload.Module.Version != info.Main.Version || payload.Module.Sum != info.Main.Sum {
		t.Fatalf("module identity = %+v", payload.Module)
	}
	if identity.CompilerSourceSHA != "UNBOUND_LOCAL_SOURCE" || identity.SourceStatus != "UNOBSERVED_VCS" {
		t.Fatalf("module metadata must not manufacture Git evidence: %+v", identity)
	}
}
