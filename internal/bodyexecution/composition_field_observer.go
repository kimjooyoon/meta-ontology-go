package bodyexecution

import (
	"fmt"
	"strings"
)

func fieldReadObserver() (string, error) {
	source, err := localReadObserver()
	if err != nil {
		return "", err
	}
	for _, edit := range [][2]string{
		{"GoooReadLocalReset();", "GoooReadLocalReset();GoooReadFieldReset();"},
		{"{GoooReadLocalSave()", "{GoooReadLocalSave();GoooReadFieldSave()"},
	} {
		if strings.Count(source, edit[0]) != 1 {
			return "", fmt.Errorf("field observer reset/save hook differs")
		}
		source = strings.Replace(source, edit[0], edit[1], 1)
	}
	return source + fieldObserver, nil
}

// Native rows contain identities and evaluated read windows, never field values.
// Writes:case,site,frame,useStart,useEnd,drops,completed,parentAssignment,completion.
// Uses:case,site,frame,latest case-relative assignment,completionClock.
// Per-field slots are copied by value at a record copy or a fixed helper call.
const fieldObserver = `
var GoooReadFieldWrites = [][9]int32{}
var GoooReadFieldUses = [][5]int32{}
var GoooReadFieldWriteStart,GoooReadFieldUseStart,GoooReadFieldFrames int
var GoooReadFieldFrameCounts,GoooReadFieldFrameDrops,GoooReadFieldWriteDrops,GoooReadFieldUseDrops [128]int
var GoooReadFieldCompletions [128]int
var GoooReadFieldReturnSeq int
var GoooReadFieldLastReturn [32]int32
type GoooReadFieldContext struct {first int;rows [32]int;seen [32]bool}
var GoooReadFieldContexts []GoooReadFieldContext
func GoooReadFieldReset() {
 GoooReadFieldWriteStart,GoooReadFieldUseStart=len(GoooReadFieldWrites),len(GoooReadFieldUses)
}
func GoooReadFieldNewSlots() [32]int32 {
 var slots [32]int32;for i:=range slots {slots[i]=-1};return slots
}
func GoooReadFieldFrame() int {
 c:=GoooReadCase
 if GoooReadFieldFrames>=65536 {GoooReadFieldFrameDrops[c]++;return -1}
 frame:=GoooReadFieldFrameCounts[c];GoooReadFieldFrameCounts[c]++;GoooReadFieldFrames++;return frame
}
func GoooReadFieldReserve(site,frame int,parent int32) int {
 c:=GoooReadCase
 if frame<0 || len(GoooReadFieldWrites)>=4096 {GoooReadFieldWriteDrops[c]++;return -1}
 slot:=len(GoooReadFieldWrites)
 GoooReadFieldWrites=append(GoooReadFieldWrites,[9]int32{int32(c),int32(site),int32(frame),int32(len(GoooReadFieldUses)-GoooReadFieldUseStart),int32(len(GoooReadFieldUses)-GoooReadFieldUseStart),int32(GoooReadFieldUseDrops[c]),0,parent,-1})
 return slot
}
func GoooReadFieldFinish(slot int,completed bool,previous *int32) {
 if slot<0 {if completed {*previous=-2};return}
 row:=&GoooReadFieldWrites[slot]
 row[4]=int32(len(GoooReadFieldUses)-GoooReadFieldUseStart)
 row[5]=int32(GoooReadFieldUseDrops[GoooReadCase])-row[5]
 if completed {row[6]=1;row[8]=int32(GoooReadFieldCompletions[GoooReadCase]);GoooReadFieldCompletions[GoooReadCase]++;*previous=int32(slot-GoooReadFieldWriteStart)}
}
func GoooReadFieldAssign[T any](site,frame int,previous *int32,evaluate func()T)(value T) {
 slot:=GoooReadFieldReserve(site,frame,-1);done:=false
 defer func(){GoooReadFieldFinish(slot,done,previous)}()
 value=evaluate();done=true;return value
}
func GoooReadFieldAssignRecord[T any](first,count,frame int,previous *[32]int32,parents [32]int32,mode int,evaluate func()T)(value T) {
 var slots [32]int
 before:=GoooReadFieldReturnSeq
 for i:=0;i<count;i++ {parent:=int32(-1);if mode==1 {parent=parents[i]};slots[i]=GoooReadFieldReserve(first+i,frame,parent)}
 context:=len(GoooReadFieldContexts)
 if mode==3 {GoooReadFieldContexts=append(GoooReadFieldContexts,GoooReadFieldContext{first:first,rows:slots})}
 done:=false
 defer func(){
  for i:=0;i<count;i++ {
   if done && mode==2 && GoooReadFieldReturnSeq!=before && slots[i]>=0 {GoooReadFieldWrites[slots[i]][7]=GoooReadFieldLastReturn[i]}
   if mode==3 {
    if slots[i]<0 {if done {previous[i]=-2};continue}
    row:=&GoooReadFieldWrites[slots[i]]
    if !GoooReadFieldContexts[context].seen[i] {row[4]=row[3];row[5]=0}
    if done {row[6]=1;row[8]=int32(GoooReadFieldCompletions[GoooReadCase]);GoooReadFieldCompletions[GoooReadCase]++;previous[i]=int32(slots[i]-GoooReadFieldWriteStart)}
   } else {GoooReadFieldFinish(slots[i],done,&previous[i])}
  }
  if mode==3 {GoooReadFieldContexts=GoooReadFieldContexts[:context]}
 }()
 value=evaluate();done=true;return value
}
func GoooReadFieldPart[T any](site int,evaluate func()T)(value T) {
 index:=len(GoooReadFieldContexts)-1
 if index<0 {return evaluate()}
 context:=&GoooReadFieldContexts[index]
 field:=site-context.first
 if field<0 || field>=32 {return evaluate()}
 context.seen[field]=true
 slot:=context.rows[field]
 if slot<0 {return evaluate()}
 start,dropped:=len(GoooReadFieldUses)-GoooReadFieldUseStart,GoooReadFieldUseDrops[GoooReadCase]
 GoooReadFieldWrites[slot][3]=int32(start)
 defer func(){row:=&GoooReadFieldWrites[slot];row[4]=int32(len(GoooReadFieldUses)-GoooReadFieldUseStart);row[5]=int32(GoooReadFieldUseDrops[GoooReadCase]-dropped)}()
 return evaluate()
}
func GoooReadFieldUse[T any](site,frame int,assignment int32,value T)T {
 if frame<0 || len(GoooReadFieldUses)>=65536 {GoooReadFieldUseDrops[GoooReadCase]++;return value}
 GoooReadFieldUses=append(GoooReadFieldUses,[5]int32{int32(GoooReadCase),int32(site),int32(frame),assignment,int32(GoooReadFieldCompletions[GoooReadCase])});return value
}
func GoooReadFieldRecordUse[T any](first,count,frame int,assignments [32]int32,value T)T {
 for i:=0;i<count;i++ {GoooReadFieldUse(first+i,frame,assignments[i],0)};return value
}
func GoooReadFieldParameter(first,count,frame int,slots *[32]int32,arguments [][32]int32,argument int) {
 for i:=0;i<count;i++ {
  parent:=int32(-1);if argument<len(arguments) {parent=arguments[argument][i]}
  row:=GoooReadFieldReserve(first+i,frame,parent);GoooReadFieldFinish(row,true,&slots[i])
 }
}
func GoooReadFieldReturn[T any](parents [32]int32,value T)T {
 GoooReadFieldLastReturn=parents;GoooReadFieldReturnSeq++;return value
}
func GoooReadFieldSave() {
 envelope:=map[string]any{"schema":"gooo/native-field-observation/v1","writes":GoooReadFieldWrites,"uses":GoooReadFieldUses,
 "frames":GoooReadFieldFrameCounts[:GoooReadCases],"frame_dropped":GoooReadFieldFrameDrops[:GoooReadCases],"write_dropped":GoooReadFieldWriteDrops[:GoooReadCases],"use_dropped":GoooReadFieldUseDrops[:GoooReadCases]}
 raw,err:=json.Marshal(envelope);if err!=nil {os.Exit(4)}
 if os.WriteFile("gooo-native-field-observation.json",raw,0600)!=nil {os.Exit(4)}
}
`
