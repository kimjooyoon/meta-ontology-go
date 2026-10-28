package syntax

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestRecordProfileContractAndSelection(t *testing.T) {
	canonical := "id=gooo.entityfields.go-projection.v7\n" +
		"types=urn:gooo:type:string,urn:gooo:type:boolean,urn:gooo:type:integer,declared-record\n" +
		"presence=required,optional\ncardinality=one,many\nordering=source\n" +
		"optional-representation=pointer\nrepeated-limit=256\n" +
		"text-bound=max_utf8_bytes\ntext-bound-unit=decoded-utf8-bytes\ntext-bound-range=0..4294967295\ntext-bound-absent=legacy\n" +
		"record-references=stable-id\nrecord-graph=acyclic\nrecord-depth-limit=8\nrecord-value-slot-limit=8192\n"
	digest := sha256.Sum256([]byte(canonical))
	if hex.EncodeToString(digest[:]) != EntityFieldsV7ProfileDigest {
		t.Fatal("V7 does not bind its complete value profile")
	}
	support := EntityFieldsV7Support()
	caps, err := support.Profile.Capabilities()
	if err != nil || !caps.Record || !caps.Many || !caps.TextBound {
		t.Fatal("record profile capabilities", caps, err)
	}
	support.Profile.Digest += "changed"
	if caps, err := support.Profile.Capabilities(); err == nil || caps != (EntityFieldsCapabilities{}) {
		t.Fatal("forged profile gained record authority")
	}
	for _, spelling := range []string{"Line", `"urn:line"`} {
		source := "package totals\nnamespace totals\nentity Basket id \"urn:basket\" fields {\nfield lines id \"urn:lines\" type " + spelling + " required many\n}\nentity Line id \"urn:line\" fields {\nfield n id \"urn:n\" type integer required one\n}\n"
		file, diagnostics := ParseFileWithEntityFieldsSupport("record.gooo", source, EntityFieldsV4Support())
		if diagnostics.HasErrors() || EntityFieldsSupportForFile(file, EntityFieldsV4Support()).Profile.Version != 7 {
			t.Fatal("forward record declaration did not select V7", diagnostics)
		}
	}
}
