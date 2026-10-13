package bodyexecution

// Typed pointers are process-local keys retaining the actual backing elements.
// They are never serialized. Keeping the keys alive prevents address reuse from
// turning a newly allocated sequence into a claim about a discarded input.
const readObserver = `
type GoooReadOrigin struct {
 Input int ` + "`json:\"input_collection\"`" + `
 Index int ` + "`json:\"element_index\"`" + `
}
type GoooReadEvent struct {
 Case int ` + "`json:\"case_index\"`" + `
 Site int ` + "`json:\"site_index\"`" + `
 Index int ` + "`json:\"index\"`" + `
 Length int ` + "`json:\"length\"`" + `
 Origin *GoooReadOrigin ` + "`json:\"origin,omitempty\"`" + `
 Reason string ` + "`json:\"reason,omitempty\"`" + `
}
type GoooReadEnvelope struct {
 Schema string ` + "`json:\"schema\"`" + `
 Reads []GoooReadEvent ` + "`json:\"reads\"`" + `
 Dropped []int ` + "`json:\"dropped\"`" + `
 RegistrationDropped []int ` + "`json:\"registration_dropped\"`" + `
}
var GoooReadCase,GoooReadCases int
var GoooReadOrigins map[any]GoooReadOrigin
var GoooReadEvents = []GoooReadEvent{}
var GoooReadDropped,GoooReadRegistrationDropped [128]int
func GoooReadReset(c,cases int) {
 GoooReadCase,GoooReadCases = c,cases
 GoooReadOrigins = make(map[any]GoooReadOrigin)
}
func GoooReadRegister[T any](values []T,input int) {
 for i := range values {
  if len(GoooReadOrigins) >= 65536 { GoooReadRegistrationDropped[GoooReadCase]++; continue }
  GoooReadOrigins[&values[i]] = GoooReadOrigin{input,i}
 }
}
func GoooReadRead[T any,I ~int|~int8|~int16|~int32|~int64|~uint|~uint8|~uint16|~uint32|~uint64|~uintptr](site int,values []T,index I) T {
 value := values[index]
 if len(GoooReadEvents) >= 65536 { GoooReadDropped[GoooReadCase]++; return value }
 event := GoooReadEvent{Case:GoooReadCase,Site:site,Index:int(index),Length:len(values)}
 if origin,found := GoooReadOrigins[&values[index]]; found { event.Origin = &origin } else { event.Reason = "NO_REGISTERED_INPUT_ELEMENT" }
 GoooReadEvents = append(GoooReadEvents,event)
 return value
}
func GoooReadSave[T any](result T) T {
 envelope:=GoooReadEnvelope{"gooo/native-read-observation/v1",GoooReadEvents,GoooReadDropped[:GoooReadCases],GoooReadRegistrationDropped[:GoooReadCases]}
 raw,err:=json.Marshal(envelope)
 if err!=nil { os.Exit(4) }
 if os.WriteFile("gooo-native-read-observation.json",raw,0600)!=nil { os.Exit(4) }
 return result
}
`
