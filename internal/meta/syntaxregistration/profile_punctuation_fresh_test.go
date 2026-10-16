package syntaxregistration

import (
	"go/format"
	"strings"
	"testing"
)

func TestProfileMigrationPunctuationFresh1859(t *testing.T) {
	for _, ending := range []string{`EntityFields:true}`, `EntityFields:true /* origin1859 */}`, `EntityFields:true,}`} {
		t.Run(ending, func(t *testing.T) {
			raw, err := format.Source([]byte(`package languagesyntax
func expectedRegistry() Registry {return Registry{Cases:[]CaseDefinition{{ID:"punctuation1859",Path:"p",` +
				ending + `}}}
`))
			if err != nil {
				t.Fatal(err)
			}
			source, err := parseSourceUnits(map[string][]byte{"punctuation1859.go": raw}, "./", "languagesyntax", false)
			if err != nil {
				t.Fatal(err)
			}
			rows, err := readProfileNativeCases(source)
			if err != nil || len(rows) != 1 {
				t.Fatalf("case origin: %v", err)
			}
			source.activity = "gooo://syntax-profile-migration/revise-native-registry"
			editProfileLiteral(source, rows[0].literal, "profile1859")
			out, err := source.finish()
			if err != nil || !strings.Contains(string(out["punctuation1859.go"]), `EntityFieldsProfile: "profile1859"`) ||
				(strings.Contains(ending, "origin1859") && !strings.Contains(string(out["punctuation1859.go"]), "origin1859")) {
				t.Fatalf("profile insertion lost punctuation or comment: %v", err)
			}
		})
	}
}
