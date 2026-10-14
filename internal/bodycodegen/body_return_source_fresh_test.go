package bodycodegen

import (
	"strconv"
	"testing"
)

func TestTypedReturnSources21641(t *testing.T) {
	const prefix = `package returnsource
namespace returnsource21641
entity Integer id "urn:gooo:type:integer"
entity Box id "urn:returnsource21641:box" fields {
 field value id "urn:returnsource21641:value" type integer required one
 field tag id "urn:returnsource21641:tag" type string required one
}
activity Make(Integer)->Box id "urn:returnsource21641:make" computes "return Box{value:input+21641,tag:\"기원\"}"
`
	for _, literal := range []string{"`return Make(input)`", strconv.Quote("return Make(input)")} {
		source := []byte(prefix + "activity Pass(Integer)->Box id \"urn:returnsource21641:pass\" computes " + literal + "\n")
		result, err := Generate("반환21641.gooo", source, "Pass")
		if err != nil {
			t.Fatal(err)
		}
		origins := result.Report.BodyReturns
		if len(origins) != 2 || result.Report.CompletenessPercent != 100 || !result.Report.RouteEquivalence.Equivalent {
			t.Fatal("typed root/helper return population", origins, result.Report)
		}
		bound, err := BindBodyReturnSources(source, origins)
		if err != nil {
			t.Fatal(err)
		}
		for i, origin := range origins {
			if origin.RecordID != "urn:returnsource21641:box" || origin.NativeOrdinal != 0 || origin.NativePopulation != 1 || origin.Value == nil || bound[i].Status != "MAPPED" {
				t.Fatal("typed authored return identity", origin, bound[i])
			}
		}
		bad := append([]BodyReturnOrigin(nil), origins...)
		bad[0].RecordID = "urn:returnsource21641:other"
		if EqualRetainedReturnOrigins(bad, origins) || !EqualRetainedReturnOrigins(nil, origins) {
			t.Fatal("explicit claims or legacy absence")
		}
	}
	t.Log("two new raw/escaped sources retain root/helper record return identities and original ranges;legacy absence supported;source-only/model0")
}
