package bodyexecution

import (
	"fmt"
	"strings"
)

// Keep the read v2 wire format intact. Local windows share its case-relative
// counters, while a separate bounded sidecar records scalar binding instances.
func localReadObserver() (string, error) {
	source := constructionReadObserver
	for _, edit := range [][2]string{
		{"GoooReadOrigins=make(map[any]GoooReadOrigin)", "GoooReadLocalReset();GoooReadOrigins=make(map[any]GoooReadOrigin)"},
		{"element:=len(GoooReadConstructions[slot].Elements)", "element:=len(GoooReadConstructions[slot].Elements);defer GoooReadLocalElement(slot,element)()"},
		{"func GoooReadSave[T any](result T) T {", "func GoooReadSave[T any](result T) T {GoooReadLocalSave()"},
	} {
		if strings.Count(source, edit[0]) != 1 {
			return "", fmt.Errorf("local observer hook differs from construction observer")
		}
		source = strings.Replace(source, edit[0], edit[1], 1)
	}
	return source + localObserver, nil
}

// Flat fixed-width rows bound native memory. No scalar values or text are copied.
// Writes: case,site,frame,readStart,readEnd,useStart,useEnd,readDrops,useDrops,done.
// Uses: case,site,frame,case-relative assignment (-1 missing, -2 write limit).
// Elements: case,construction,element,useStart,useEnd,useDrops.
const localObserver = `
var GoooReadLocalWrites = [][10]int{}
var GoooReadLocalUses = [][4]int{}
var GoooReadLocalElements = [][6]int{}
var GoooReadLocalWriteStart,GoooReadLocalUseStart,GoooReadLocalFrames int
var GoooReadLocalFrameCounts,GoooReadLocalFrameDrops,GoooReadLocalWriteDrops,GoooReadLocalUseDrops [128]int
func GoooReadLocalReset() {
 GoooReadLocalWriteStart,GoooReadLocalUseStart=len(GoooReadLocalWrites),len(GoooReadLocalUses)
}
func GoooReadLocalFrame() int {
 c:=GoooReadCase
 if GoooReadLocalFrames>=65536 {GoooReadLocalFrameDrops[c]++;return -1}
 frame:=GoooReadLocalFrameCounts[c];GoooReadLocalFrameCounts[c]++;GoooReadLocalFrames++
 return frame
}
func GoooReadLocalAssign[T any](site,frame int,previous *int,evaluate func() T) (value T) {
 c:=GoooReadCase
 if frame<0 || len(GoooReadLocalWrites)>=4096 {
  GoooReadLocalWriteDrops[c]++;value=evaluate();*previous=-2;return value
 }
 slot:=len(GoooReadLocalWrites)
 rd,ud:=GoooReadDropped[c],GoooReadLocalUseDrops[c]
 GoooReadLocalWrites=append(GoooReadLocalWrites,[10]int{c,site,frame,len(GoooReadEvents)-GoooReadCaseStart,0,len(GoooReadLocalUses)-GoooReadLocalUseStart})
 defer func(){
  row:=&GoooReadLocalWrites[slot];row[4]=len(GoooReadEvents)-GoooReadCaseStart;row[6]=len(GoooReadLocalUses)-GoooReadLocalUseStart
  row[7]=GoooReadDropped[c]-rd;row[8]=GoooReadLocalUseDrops[c]-ud
 }()
 value=evaluate()
 GoooReadLocalWrites[slot][9]=1;*previous=slot-GoooReadLocalWriteStart
 return value
}
func GoooReadLocalUse[T any](site,frame,assignment int,value T) T {
 if frame<0 || len(GoooReadLocalUses)>=65536 {GoooReadLocalUseDrops[GoooReadCase]++;return value}
 GoooReadLocalUses=append(GoooReadLocalUses,[4]int{GoooReadCase,site,frame,assignment})
 return value
}
func GoooReadLocalElement(slot,element int) func() {
 c,instance:=GoooReadCase,slot-GoooReadConstructionStart
 start,dropped:=len(GoooReadLocalUses)-GoooReadLocalUseStart,GoooReadLocalUseDrops[c]
 return func(){GoooReadLocalElements=append(GoooReadLocalElements,[6]int{c,instance,element,start,len(GoooReadLocalUses)-GoooReadLocalUseStart,GoooReadLocalUseDrops[c]-dropped})}
}
func GoooReadLocalSave() {
 envelope:=map[string]any{"schema":"gooo/native-local-observation/v1","writes":GoooReadLocalWrites,"uses":GoooReadLocalUses,"elements":GoooReadLocalElements,
 "frames":GoooReadLocalFrameCounts[:GoooReadCases],"frame_dropped":GoooReadLocalFrameDrops[:GoooReadCases],"write_dropped":GoooReadLocalWriteDrops[:GoooReadCases],"use_dropped":GoooReadLocalUseDrops[:GoooReadCases]}
 raw,err:=json.Marshal(envelope);if err!=nil {os.Exit(4)}
 if os.WriteFile("gooo-native-local-observation.json",raw,0600)!=nil {os.Exit(4)}
}
`
