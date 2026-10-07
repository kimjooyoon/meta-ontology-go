package buildidentity

import (
	"encoding/json"
	"reflect"
	"runtime/debug"
	"strings"
	"testing"
)

func TestModuleFromBuildInfoPreservesObservedCoordinates(t *testing.T) {
	info := &debug.BuildInfo{Main: debug.Module{Path: "example.org/compiler", Version: "v1.2.3", Sum: "h1:observed"}}
	want := &Module{Path: info.Main.Path, Version: info.Main.Version, Sum: info.Main.Sum}
	if got := ModuleFromBuildInfo(info); !reflect.DeepEqual(got, want) {
		t.Fatalf("module = %+v, want %+v", got, want)
	}
	for _, absent := range []*debug.BuildInfo{nil, {}, {Main: debug.Module{Version: "v1.2.3"}}} {
		if got := ModuleFromBuildInfo(absent); got != nil {
			t.Fatalf("unobserved module path = %+v", got)
		}
	}
}

func TestModuleFromBuildInfoRetainsReplacementWithoutAliasing(t *testing.T) {
	for _, replacement := range []debug.Module{
		{Path: "example.org/fork", Version: "v2.0.0", Sum: "h1:replacement"},
		{Path: "../local-compiler"},
	} {
		info := &debug.BuildInfo{Main: debug.Module{Path: "example.org/compiler", Version: "v1.2.3", Replace: &replacement}}
		got := ModuleFromBuildInfo(info)
		want := &ModuleReplacement{Path: replacement.Path, Version: replacement.Version, Sum: replacement.Sum}
		info.Main.Path, replacement.Path = "changed", "changed"
		if got.Path != "example.org/compiler" || !reflect.DeepEqual(got.Replacement, want) {
			t.Fatalf("replacement snapshot changed: %+v", got)
		}
	}
}

func TestModuleFromBuildInfoDoesNotInventMissingVersionOrChecksum(t *testing.T) {
	for _, version := range []string{"", "(devel)"} {
		got := ModuleFromBuildInfo(&debug.BuildInfo{Main: debug.Module{Path: "example.org/compiler", Version: version}})
		raw, err := json.Marshal(got)
		if err != nil {
			t.Fatal(err)
		}
		if got.Version != version || got.Sum != "" || strings.Contains(string(raw), `"sum"`) {
			t.Fatalf("unobserved checksum or version changed: %s", raw)
		}
	}
}
