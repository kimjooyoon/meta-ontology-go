package bodyexecution

import "strings"

// The predicate observes every input element. Only retained values receive a
// construction element; its read interval includes the source element read.
func withCollectionFilterObserver(observer string) string {
	source := collectionFilterReadObserver
	if strings.Contains(observer, "func GoooReadLocalElement(") {
		source = strings.Replace(source, "  if predicate(value) {\n   values[length]=value", "  finish:=GoooReadLocalElement(slot,length)\n  if predicate(value) {\n   values[length]=value", 1)
		source = strings.Replace(source, "   count=length\n", "   count=length\n   finish()\n", 1)
	}
	return withCollectionCountField(observer) + source
}

const collectionFilterReadObserver = `
func GoooReadFilter[T any](site,readSite int,input []T,predicate func(T) bool) []T {
 if len(input)>256 {panic("collection filter exceeds 256 elements")}
 var values [256]T
 length:=0
 if len(GoooReadConstructions)>=4096 || GoooReadReserved+len(input)>65536 {
  GoooReadConstructionDropped[GoooReadCase]++
  for i:=range input {
   value:=GoooReadRead(readSite,input,i)
   if predicate(value) {values[length]=value;length++}
  }
  return values[:length:length]
 }
 slot:=len(GoooReadConstructions)
 instance:=slot-GoooReadConstructionStart
 GoooReadReserved+=len(input)
 count:=0
 GoooReadConstructions=append(GoooReadConstructions,GoooReadConstruction{
  Case:GoooReadCase,Site:site,ReadStart:len(GoooReadEvents)-GoooReadCaseStart,
  Elements:make([]GoooReadElement,0,len(input)),ElementCount:&count,
 })
 defer func(){GoooReadConstructions[slot].ReadEnd=len(GoooReadEvents)-GoooReadCaseStart}()
 for i:=range input {
  start:=len(GoooReadEvents)-GoooReadCaseStart
  dropped:=GoooReadDropped[GoooReadCase]
  value:=GoooReadRead(readSite,input,i)
  if predicate(value) {
   values[length]=value
   inputIndex:=i
   GoooReadConstructions[slot].Elements=append(GoooReadConstructions[slot].Elements,GoooReadElement{
    ReadStart:start,ReadEnd:len(GoooReadEvents)-GoooReadCaseStart,
    ReadsDropped:GoooReadDropped[GoooReadCase]-dropped,Completed:true,
    InputIndex:&inputIndex,
   })
   length++
   count=length
  }
 }
 GoooReadConstructions[slot].Completed=true
 output:=values[:length:length]
 for i:=range output {GoooReadConstructed[&output[i]]=GoooReadConstructedOrigin{instance,i}}
 return output
}
`
