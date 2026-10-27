package main

const originNativeConsumer = `package origins
import (
 "encoding/json"
 "os"
 "testing"
)
func TestReadOriginWithoutEvaluatingZeroInput(t *testing.T) {
 raw,err:=os.ReadFile("expected.json"); if err!=nil {t.Fatal(err)}
 if GoooGenerationOrigin!=string(raw) {t.Fatal("origin changed in importing module")}
 var origin map[string]any
 if err:=json.Unmarshal([]byte(GoooGenerationOrigin),&origin);err!=nil {t.Fatal(err)}
 if origin["schema"]!="gooo/library-generation-origin/v1" || origin["entry_activity_id"]!="urn:origin:divide" {t.Fatal(origin)}
 for _,row:=range []struct{input,want int64}{{2,5},{5,2}} {
  if got:=Evaluate(R_Value{F_number:row.input});got.F_number!=row.want {t.Fatal(got,row)}
 }
}
`
