package bodyexecution

import "strings"

// Only dynamic constructors add a count field. Literal-only projections retain
// their original observer source and receipt hashes.
func withCollectionMapObserver(observer string) string {
	return withCollectionCountField(observer) + collectionMapReadObserver
}

func withCollectionCountField(observer string) string {
	if strings.Contains(observer, "ElementCount *int") {
		return observer
	}
	field := "\nElementCount *int `json:\"element_count,omitempty\"`\n"
	return strings.Replace(observer, "type GoooReadConstruction struct {", "type GoooReadConstruction struct {"+field, 1)
}

const collectionMapReadObserver = `
func GoooReadMap[In,Out any](site,readSite int,input []In,mapper func(In) Out) []Out {
 if len(input)>256 {panic("collection map exceeds 256 elements")}
 values:=make([]Out,len(input))
 if len(GoooReadConstructions)>=4096 || GoooReadReserved+len(input)>65536 {
  GoooReadConstructionDropped[GoooReadCase]++
  for i:=range input {values[i]=mapper(GoooReadRead(readSite,input,i))}
  return values
 }
 slot:=len(GoooReadConstructions)
 instance:=slot-GoooReadConstructionStart
 count:=len(input)
 GoooReadReserved+=len(input)
 GoooReadConstructions=append(GoooReadConstructions,GoooReadConstruction{
  Case:GoooReadCase,Site:site,ReadStart:len(GoooReadEvents)-GoooReadCaseStart,
  Elements:make([]GoooReadElement,0,len(input)),ElementCount:&count,
 })
 defer func(){GoooReadConstructions[slot].ReadEnd=len(GoooReadEvents)-GoooReadCaseStart}()
 for i:=range input {
  values[i]=GoooReadEvaluate(slot,func() Out {return mapper(GoooReadRead(readSite,input,i))})
 }
 GoooReadConstructions[slot].Completed=true
 for i:=range values {GoooReadConstructed[&values[i]]=GoooReadConstructedOrigin{instance,i}}
 return values
}
`
