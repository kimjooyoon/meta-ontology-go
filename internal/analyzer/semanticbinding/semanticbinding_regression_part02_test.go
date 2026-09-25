package semanticbinding

import (
	"bytes"
	"testing"
)

func TestFilenameChangeRetainsIdentityAndOnlyChangesLocation(t *testing.T) {
	aSource, aWant := loadFixture(t, "filename_identity_a")
	bSource, bWant := loadFixture(t, "filename_identity_b")
	if !bytes.Equal(aSource, bSource) {
		t.Fatal("filename fixtures must have identical source bytes")
	}
	if len(aWant.Records) != 1 || len(bWant.Records) != 1 {
		t.Fatalf("filename records = %#v and %#v, want one record each", aWant.Records, bWant.Records)
	}
	aRecord, bRecord := aWant.Records[0], bWant.Records[0]
	if aRecord.Directive != bRecord.Directive || aRecord.Name != bRecord.Name ||
		aRecord.ID != bRecord.ID || aRecord.Subject != bRecord.Subject || aRecord.Pressure != bRecord.Pressure {
		t.Fatalf("filename changed semantic record: a=%#v b=%#v", aRecord, bRecord)
	}
	if aRecord.Location == nil || bRecord.Location == nil || aRecord.Location.Filename == bRecord.Location.Filename {
		t.Fatalf("filename locations were not retained independently: a=%#v b=%#v", aRecord.Location, bRecord.Location)
	}
	if aRecord.Location.Start != bRecord.Location.Start || aRecord.Location.End != bRecord.Location.End {
		t.Fatalf("filename changed non-location span data: a=%#v b=%#v", aRecord.Location, bRecord.Location)
	}
}
