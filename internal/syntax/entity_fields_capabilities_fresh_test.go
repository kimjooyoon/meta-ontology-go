package syntax

import "testing"

func TestExactFieldCapabilitiesFresh1905(t *testing.T) {
	profiles := []EntityFieldsSupport{EntityFieldsV1Support(), EntityFieldsV2Support(), EntityFieldsV3Support(),
		EntityFieldsV4Support(), EntityFieldsV5Support(), EntityFieldsV6Support()}
	for i, support := range profiles {
		capabilities, err := support.Profile.Capabilities()
		want := EntityFieldsCapabilities{Boolean: i >= 1, Integer: i >= 2, Optional: i >= 3, Many: i >= 4, TextBound: i >= 5}
		if err != nil || capabilities != want {
			t.Fatal("profile capability changed", i+1, capabilities, want, err)
		}
		for _, field := range []string{"identity", "version", "digest"} {
			forged := support.Profile
			switch field {
			case "identity":
				forged.ID = "gooo.entityfields.unknown"
			case "version":
				forged.Version = 1000000
			case "digest":
				forged.Digest = EntityFieldsV6ProfileDigest + "changed"
			}
			if got, err := forged.Capabilities(); err == nil || got != (EntityFieldsCapabilities{}) {
				t.Fatal("invalid profile granted a capability", forged, got, err)
			}
		}
	}
}

func TestSourceCannotRepairUnboundProfileFresh1905(t *testing.T) {
	file, diagnostics := ParseFileWithEntityFieldsSupport("bound.gooo", textBoundSource1905, EntityFieldsV6Support())
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	forged := EntityFieldsV4Support()
	forged.Profile.Digest = "different"
	for _, prior := range []EntityFieldsSupport{CurrentEntityFieldsSupport(), forged, {State: "UNKNOWN"}} {
		if got := EntityFieldsSupportForFile(file, prior); got != prior {
			t.Fatal("source repaired or activated an unavailable caller profile", prior, got)
		}
	}
}
