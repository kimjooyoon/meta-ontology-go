package syntax

import "testing"

func TestSupportedEntityFieldsCatalogPreservesIdentityAndDetachedValues(t *testing.T) {
	profiles := SupportedEntityFieldsProfiles()
	want := []string{EntityFieldsProfileID, EntityFieldsV2ProfileID, EntityFieldsV3ProfileID,
		EntityFieldsV4ProfileID, EntityFieldsV5ProfileID, EntityFieldsV6ProfileID, EntityFieldsV7ProfileID}
	if len(profiles) != len(want) {
		t.Fatal("kernel profile catalog changed", len(profiles))
	}
	for i, p := range profiles {
		if p.ID != want[i] || p.Version != i+1 || p.Validate() != nil {
			t.Fatalf("catalog identity %d: %+v", i, p)
		}
		resolved, ok := EntityFieldsSupportForProfileID(want[i])
		if !ok || resolved.Profile != p {
			t.Fatal("canonical lookup differs from supported profile", p)
		}
	}
	profiles[6].Digest = "mutated"
	profiles[0].ID = "invented"
	if SupportedEntityFieldsProfiles()[6].Validate() != nil ||
		SupportedEntityFieldsProfiles()[0].ID != EntityFieldsProfileID {
		t.Fatal("caller changed the kernel profile table")
	}
}
