package main

// This consumer imports ordinary generated packages. Its expected results come
// from the existing independent example corpus and handwritten arithmetic.
const libraryConsumerOracle = `package libraryprobe_test
import (
 "encoding/json"
 "os"
 "reflect"
 "testing"
 "libraryprobe/surface"
 "libraryprobe/optional"
 "libraryprobe/named"
 "libraryprobe/hygiene"
)
func check[I, O any](t *testing.T, file, input, entry string, count int, evaluate func(I) O) {
 t.Helper()
 raw, err := os.ReadFile(file)
 if err != nil { t.Fatal(err) }
 var suite struct { Cases []struct { Inputs, Expected map[string]json.RawMessage } }
 if err := json.Unmarshal(raw, &suite); err != nil { t.Fatal(err) }
 if len(suite.Cases) != count { t.Fatal("independent case population changed") }
 for i, row := range suite.Cases {
  var in I
  var want O
  if err := json.Unmarshal(row.Inputs[input], &in); err != nil { t.Fatal(err) }
  if err := json.Unmarshal(row.Expected[entry], &want); err != nil { t.Fatal(err) }
  actual := evaluate(in)
  if !reflect.DeepEqual(actual, want) { t.Fatalf("case %d got %#v want %#v", i, actual, want) }
 }
}
func TestIndependentSurface(t *testing.T) {
 check(t, "surface.json", "example/runtime-package-surface:Inspect", "example/runtime-package-surface:Inspect", 14, surface.Evaluate)
 field, ok := reflect.TypeFor[surface.R_Observation]().FieldByName("F_origin_ref")
 if !ok || field.Tag.Get("gooo") != "urn:gooo:runtime-package-surface:observation:origin" ||
  field.Tag.Get("json") != "origin_ref" { t.Fatal("field lost source identity", field) }
}
func TestIndependentOptional(t *testing.T) {
 check(t, "optional.json", "example/domain:Submit", "example/app:Main", 3, optional.Evaluate)
 empty, zero, no := "", int64(0), false
 out := optional.Evaluate(optional.R_Profile{F_note: &empty, F_count: &zero, F_complete: &no})
 raw, err := json.Marshal(out)
 if err != nil || string(raw) != "{\"note\":\"\",\"complete\":false,\"count\":0}" { t.Fatal(string(raw), err) }
 absent, err := json.Marshal(optional.Evaluate(optional.R_Profile{}))
 if err != nil || string(absent) != "{}" { t.Fatal(string(absent), err) }
}
func TestIndependentOrderedRoots(t *testing.T) {
 for _, row := range [][3]int64{{11,5,29},{-37,5,-19},{0,23,0}} {
  if got := named.Evaluate(row[0],row[1]); got != row[2] { t.Fatal(row,got) }
 }
}
func TestIndependentHygiene(t *testing.T) {
 if got := hygiene.Evaluate(hygiene.R_Record{F_label:"before"}); got.F_label != "before|before|literal" {
  t.Fatal("public types or private helpers captured source locals", got)
 }
}
`

const libraryHygieneSource = `package hygiene
namespace hygiene
entity Record id "urn:library:record" fields {
 field label id "urn:library:label" type string required one
}
activity Copy(Record) -> Record id "urn:library:copy" computes "return input"
activity Keep(Record) -> Record id "urn:library:keep" computes ` + "`" + `
let R_Record = input.label
let goooComposedActivity0Call0 = R_Record
let goooLibraryLocal0 = "|literal"
return Copy(Record{label: R_Record + "|" + goooComposedActivity0Call0 + goooLibraryLocal0})
` + "`" + `
`
