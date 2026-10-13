package bodyexecution

// Construction pointers retain each actual backing array until its case finishes.
// Element closures execute once in source order; deferred observations do not recover faults.
const constructionReadObserver = `
type GoooReadOrigin struct { Input int ` + "`" + `json:"input_collection"` + "`" + `; Index int ` + "`" + `json:"element_index"` + "`" + ` }
type GoooReadConstructedOrigin struct { Instance int ` + "`" + `json:"instance_index"` + "`" + `; Index int ` + "`" + `json:"element_index"` + "`" + ` }
type GoooReadEvent struct {
 Case int ` + "`" + `json:"case_index"` + "`" + `
 Site int ` + "`" + `json:"site_index"` + "`" + `
 Index int ` + "`" + `json:"index"` + "`" + `
 Length int ` + "`" + `json:"length"` + "`" + `
 Origin *GoooReadOrigin ` + "`" + `json:"origin,omitempty"` + "`" + `
 Construction *GoooReadConstructedOrigin ` + "`" + `json:"construction,omitempty"` + "`" + `
 Reason string ` + "`" + `json:"reason,omitempty"` + "`" + `
}
type GoooReadElement struct {
 ReadStart int ` + "`" + `json:"read_start"` + "`" + `
 ReadEnd int ` + "`" + `json:"read_end"` + "`" + `
 ReadsDropped int ` + "`" + `json:"reads_dropped"` + "`" + `
 Completed bool ` + "`" + `json:"completed"` + "`" + `
}
type GoooReadConstruction struct {
 Case int ` + "`" + `json:"case_index"` + "`" + `
 Site int ` + "`" + `json:"site_index"` + "`" + `
 ReadStart int ` + "`" + `json:"read_start"` + "`" + `
 ReadEnd int ` + "`" + `json:"read_end"` + "`" + `
 Completed bool ` + "`" + `json:"completed"` + "`" + `
 Elements []GoooReadElement ` + "`" + `json:"elements"` + "`" + `
}
type GoooReadEnvelope struct {
 Schema string ` + "`" + `json:"schema"` + "`" + `
 Reads []GoooReadEvent ` + "`" + `json:"reads"` + "`" + `
 Dropped []int ` + "`" + `json:"dropped"` + "`" + `
 RegistrationDropped []int ` + "`" + `json:"registration_dropped"` + "`" + `
 Constructions []GoooReadConstruction ` + "`" + `json:"constructions"` + "`" + `
 ConstructionDropped []int ` + "`" + `json:"construction_dropped"` + "`" + `
}
var GoooReadCase,GoooReadCases,GoooReadCaseStart,GoooReadConstructionStart,GoooReadReserved int
var GoooReadOrigins map[any]GoooReadOrigin
var GoooReadConstructed map[any]GoooReadConstructedOrigin
var GoooReadEvents = []GoooReadEvent{}
var GoooReadConstructions = []GoooReadConstruction{}
var GoooReadDropped,GoooReadRegistrationDropped,GoooReadConstructionDropped [128]int
func GoooReadReset(c,cases int) {
 GoooReadCase,GoooReadCases=c,cases
 GoooReadCaseStart,GoooReadConstructionStart=len(GoooReadEvents),len(GoooReadConstructions)
 GoooReadOrigins=make(map[any]GoooReadOrigin)
 GoooReadConstructed=make(map[any]GoooReadConstructedOrigin)
}
func GoooReadRegister[T any](values []T,input int) {
 for i:=range values {
  if len(GoooReadOrigins)>=65536 {GoooReadRegistrationDropped[GoooReadCase]++;continue}
  GoooReadOrigins[&values[i]]=GoooReadOrigin{input,i}
 }
}
func GoooReadRead[T any,I ~int|~int8|~int16|~int32|~int64|~uint|~uint8|~uint16|~uint32|~uint64|~uintptr](site int,values []T,index I) T {
 value:=values[index]
 if len(GoooReadEvents)>=65536 {GoooReadDropped[GoooReadCase]++;return value}
 event:=GoooReadEvent{Case:GoooReadCase,Site:site,Index:int(index),Length:len(values)}
 if origin,found:=GoooReadOrigins[&values[index]];found {event.Origin=&origin
 } else if origin,found:=GoooReadConstructed[&values[index]];found {event.Construction=&origin
 } else {event.Reason="NO_REGISTERED_INPUT_ELEMENT"}
 GoooReadEvents=append(GoooReadEvents,event)
 return value
}
func GoooReadMake[T any](site int,elements ...func() T) []T {
 values:=make([]T,len(elements))
 if len(GoooReadConstructions)>=4096 || GoooReadReserved+len(elements)>65536 {
  GoooReadConstructionDropped[GoooReadCase]++
  for i,e:=range elements {values[i]=e()}
  return values
 }
 slot:=len(GoooReadConstructions)
 instance:=slot-GoooReadConstructionStart
 GoooReadReserved+=len(elements)
 GoooReadConstructions=append(GoooReadConstructions,GoooReadConstruction{Case:GoooReadCase,Site:site,ReadStart:len(GoooReadEvents)-GoooReadCaseStart,Elements:[]GoooReadElement{}})
 defer func(){GoooReadConstructions[slot].ReadEnd=len(GoooReadEvents)-GoooReadCaseStart}()
 for i,e:=range elements {values[i]=GoooReadEvaluate(slot,e)}
 GoooReadConstructions[slot].Completed=true
 for i:=range values {GoooReadConstructed[&values[i]]=GoooReadConstructedOrigin{instance,i}}
 return values
}
func GoooReadEvaluate[T any](slot int,evaluate func() T) (value T) {
 element:=len(GoooReadConstructions[slot].Elements)
 dropped:=GoooReadDropped[GoooReadCase]
 GoooReadConstructions[slot].Elements=append(GoooReadConstructions[slot].Elements,GoooReadElement{ReadStart:len(GoooReadEvents)-GoooReadCaseStart})
 defer func(){
  item:=&GoooReadConstructions[slot].Elements[element]
  item.ReadEnd=len(GoooReadEvents)-GoooReadCaseStart
  item.ReadsDropped=GoooReadDropped[GoooReadCase]-dropped
 }()
 value=evaluate()
 GoooReadConstructions[slot].Elements[element].Completed=true
 return value
}
func GoooReadSave[T any](result T) T {
 envelope:=GoooReadEnvelope{Schema:"gooo/native-read-observation/v2",Reads:GoooReadEvents,Dropped:GoooReadDropped[:GoooReadCases],RegistrationDropped:GoooReadRegistrationDropped[:GoooReadCases],Constructions:GoooReadConstructions,ConstructionDropped:GoooReadConstructionDropped[:GoooReadCases]}
 raw,err:=json.Marshal(envelope);if err!=nil {os.Exit(4)}
 if os.WriteFile("gooo-native-read-observation.json",raw,0600)!=nil {os.Exit(4)}
 return result
}
`
