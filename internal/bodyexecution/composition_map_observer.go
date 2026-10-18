package bodyexecution

// The collection-map projection will opt into this kernel after source typing,
// call-budget accounting and dynamic construction-site binding are implemented.
// Keeping it separate preserves existing native projection bytes in the meantime.
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
 GoooReadReserved+=len(input)
 GoooReadConstructions=append(GoooReadConstructions,GoooReadConstruction{
  Case:GoooReadCase,Site:site,ReadStart:len(GoooReadEvents)-GoooReadCaseStart,
  Elements:make([]GoooReadElement,0,len(input)),
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
