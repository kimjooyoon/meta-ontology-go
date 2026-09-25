package languagediagnosticprovenance

import "testing"

func TestUnknownStageLowersDiagnosticResolution(t *testing.T) {
	observation, found := sourceMapObservation("entity")
	if !found {
		t.Fatal("source-map fixture missing")
	}
	observation.Stage = "UNKNOWN"
	_, failure := Normalize(observation)
	if failure == nil || failure.Code != "PROVENANCE_STAGE_UNKNOWN" {
		t.Fatalf("unknown stage = %#v", failure)
	}
}

func TestAmbiguousSourceMapFailsClosed(t *testing.T) {
	observation, found := sourceMapObservation("field")
	if !found {
		t.Fatal("source-map fixture missing")
	}
	observation.SourceMap.Mappings = append(
		observation.SourceMap.Mappings,
		observation.SourceMap.Mappings[0],
	)
	_, failure := Normalize(observation)
	if failure == nil || failure.Code != "SOURCE_MAP_AMBIGUOUS" {
		t.Fatalf("ambiguous source map = %#v", failure)
	}
}
