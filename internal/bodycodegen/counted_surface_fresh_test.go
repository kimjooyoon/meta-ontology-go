package bodycodegen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestScopedCounterSurface18121(t *testing.T) {
	const header = `package loopform
namespace loopform18121
entity Integer id "urn:gooo:type:integer"
`
	const body = `let total=input+18121
for let i=0;i<4;i++ {total=total+i}
for let i=0;i<3;i++ {total=total+i*2}
let i=37
return total+i`
	var reports []Report
	for _, literal := range []string{"`" + body + "`", strconv.Quote(body)} {
		source := []byte(header + `activity Sum(Integer)->Integer id "urn:loopform18121:sum" computes ` + literal + "\n")
		result, err := Generate("counter18121.gooo", source, "Sum")
		if err != nil {
			t.Fatal(err)
		}
		if result.Report.CompletenessPercent != 100 || !result.Report.RouteEquivalence.Equivalent {
			t.Fatal("lowering units or equivalence", result.Report)
		}
		origins := result.Report.BodyLocals
		bound, err := BindBodyLocalSources(source, origins)
		if err != nil {
			t.Fatal(err)
		}
		variables := map[int]bool{}
		updates := 0
		for j, origin := range origins {
			if origin.Name == "i" {
				variables[origin.VariableOrdinal] = true
			}
			if origin.Operation != "++" {
				continue
			}
			updates++
			if origin.Kind != "update" || bound[j].Value.Status != "MAPPED" || bound[j].Value.Expression != "i++" {
				t.Fatal("authored increment lost", origin, bound[j])
			}
		}
		uses := 0
		for _, b := range bound {
			for _, use := range b.Uses {
				if use.Source != nil && use.Source.Status == "MAPPED" && use.Source.Expression == "i" {
					uses++
				}
			}
		}
		if len(variables) != 3 || updates != 2 || uses < 4 {
			t.Fatal("scoped counter identities/read origins", variables, updates, uses)
		}
		reports = append(reports, result.Report)
	}
	if root := os.Getenv("GOOO_FRESH_SOURCE_OUTPUT"); root != "" {
		raw, err := json.MarshalIndent(reports, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "counter18121.json"), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("raw/escaped new source;two scoped counters plus later same-name local remain three identities;four authored increment slots retain ++ and implicit reads;canonical lowering100%;producer/model0")
}
