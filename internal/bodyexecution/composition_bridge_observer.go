package bodyexecution

import (
	"fmt"
	"strings"
)

// Bridge windows join the retained local and field journals by their own
// case-relative row numbers. Neither values nor frame-number equality is used.
func bridgeReadObserver(source string) (string, error) {
	groups := [][][2]string{bridgeLocalHooks(), bridgeFieldHooks(), bridgePartHooks()}
	for group, edits := range groups {
		for index, edit := range edits {
			if strings.Count(source, edit[0]) != 1 {
				return "", fmt.Errorf("field/local bridge hook differs: group%d hook%d", group, index)
			}
			source = strings.Replace(source, edit[0], edit[1], 1)
		}
	}
	return source + bridgeObserver, nil
}

func bridgeLocalHooks() [][2]string {
	return [][2]string{
		{`GoooReadLocalWrites=append(GoooReadLocalWrites,[10]int{c,site,frame,len(GoooReadEvents)-GoooReadCaseStart,0,len(GoooReadLocalUses)-GoooReadLocalUseStart})`,
			`GoooReadLocalWrites=append(GoooReadLocalWrites,[10]int{c,site,frame,len(GoooReadEvents)-GoooReadCaseStart,0,len(GoooReadLocalUses)-GoooReadLocalUseStart});GoooReadBridgeLocalStart(slot)`},
		{`row[7]=GoooReadDropped[c]-rd;row[8]=GoooReadLocalUseDrops[c]-ud`,
			`row[7]=GoooReadDropped[c]-rd;row[8]=GoooReadLocalUseDrops[c]-ud;GoooReadBridgeLocalFinish(slot)`},
		{`GoooReadLocalUses=append(GoooReadLocalUses,[4]int{GoooReadCase,site,frame,assignment})`,
			`GoooReadLocalUses=append(GoooReadLocalUses,[4]int{GoooReadCase,site,frame,assignment});GoooReadBridgeLocalClocks=append(GoooReadBridgeLocalClocks,GoooReadBridgeCompletions[GoooReadCase])`},
	}
}

func bridgeFieldHooks() [][2]string {
	return [][2]string{
		{`GoooReadFieldWrites=append(GoooReadFieldWrites,[9]int32{int32(c),int32(site),int32(frame),int32(len(GoooReadFieldUses)-GoooReadFieldUseStart),int32(len(GoooReadFieldUses)-GoooReadFieldUseStart),int32(GoooReadFieldUseDrops[c]),0,parent,-1})`,
			`GoooReadFieldWrites=append(GoooReadFieldWrites,[9]int32{int32(c),int32(site),int32(frame),int32(len(GoooReadFieldUses)-GoooReadFieldUseStart),int32(len(GoooReadFieldUses)-GoooReadFieldUseStart),int32(GoooReadFieldUseDrops[c]),0,parent,-1});GoooReadBridgeFieldStart(slot)`},
		{`if completed {row[6]=1;row[8]=int32(GoooReadFieldCompletions[GoooReadCase]);GoooReadFieldCompletions[GoooReadCase]++;*previous=int32(slot-GoooReadFieldWriteStart)}`,
			`if completed {row[6]=1;row[8]=int32(GoooReadFieldCompletions[GoooReadCase]);GoooReadFieldCompletions[GoooReadCase]++;*previous=int32(slot-GoooReadFieldWriteStart)};GoooReadBridgeFieldFinish(slot,completed,false)`},
		{`if !GoooReadFieldContexts[context].seen[i] {row[4]=row[3];row[5]=0}`,
			`if !GoooReadFieldContexts[context].seen[i] {row[4]=row[3];row[5]=0;GoooReadBridgeFieldEmpty(slots[i])}`},
		{`if done {row[6]=1;row[8]=int32(GoooReadFieldCompletions[GoooReadCase]);GoooReadFieldCompletions[GoooReadCase]++;previous[i]=int32(slots[i]-GoooReadFieldWriteStart)}`,
			`if done {row[6]=1;row[8]=int32(GoooReadFieldCompletions[GoooReadCase]);GoooReadFieldCompletions[GoooReadCase]++;previous[i]=int32(slots[i]-GoooReadFieldWriteStart)};GoooReadBridgeFieldFinish(slots[i],done,true)`},
		{`GoooReadFieldUses=append(GoooReadFieldUses,[5]int32{int32(GoooReadCase),int32(site),int32(frame),assignment,int32(GoooReadFieldCompletions[GoooReadCase])});return value`,
			`GoooReadFieldUses=append(GoooReadFieldUses,[5]int32{int32(GoooReadCase),int32(site),int32(frame),assignment,int32(GoooReadFieldCompletions[GoooReadCase])});GoooReadBridgeFieldClocks=append(GoooReadBridgeFieldClocks,GoooReadBridgeCompletions[GoooReadCase]);return value`},
		{`func GoooReadFieldSave() {`, `func GoooReadFieldSave() {GoooReadBridgeSave()`},
	}
}

func bridgePartHooks() [][2]string {
	return [][2]string{
		{`GoooReadFieldWrites[slot][3]=int32(start)`, `GoooReadFieldWrites[slot][3]=int32(start);GoooReadBridgePartStart(slot)`},
		{`defer func(){row:=&GoooReadFieldWrites[slot];row[4]=int32(len(GoooReadFieldUses)-GoooReadFieldUseStart);row[5]=int32(GoooReadFieldUseDrops[GoooReadCase]-dropped)}()`,
			`defer func(){row:=&GoooReadFieldWrites[slot];row[4]=int32(len(GoooReadFieldUses)-GoooReadFieldUseStart);row[5]=int32(GoooReadFieldUseDrops[GoooReadCase]-dropped);GoooReadBridgePartEnd(slot)}()`},
	}
}

// Native bridge rows: case, own assignment, other-use start/end, other-use drops,
// shared clock at evaluation start, shared completion (-1 if unfinished).
// Every row/read corresponds one-for-one to a retained original journal row.
const bridgeObserver = `
var GoooReadBridgeLocals,GoooReadBridgeFields [][7]int32
var GoooReadBridgeLocalClocks,GoooReadBridgeFieldClocks []int32
var GoooReadBridgeCompletions [128]int32
func GoooReadBridgeLocalStart(slot int) {
 c:=GoooReadCase
 start:=int32(len(GoooReadFieldUses)-GoooReadFieldUseStart)
 GoooReadBridgeLocals=append(GoooReadBridgeLocals,[7]int32{int32(c),int32(slot-GoooReadLocalWriteStart),start,start,int32(GoooReadFieldUseDrops[c]),GoooReadBridgeCompletions[c],-1})
}
func GoooReadBridgeLocalFinish(slot int) {
 c:=GoooReadCase;row:=&GoooReadBridgeLocals[slot]
 row[3]=int32(len(GoooReadFieldUses)-GoooReadFieldUseStart)
 row[4]=int32(GoooReadFieldUseDrops[c])-row[4]
 if GoooReadLocalWrites[slot][9]==1 {row[6]=GoooReadBridgeCompletions[c];GoooReadBridgeCompletions[c]++}
}
func GoooReadBridgeFieldStart(slot int) {
 c:=GoooReadCase
 start:=int32(len(GoooReadLocalUses)-GoooReadLocalUseStart)
 GoooReadBridgeFields=append(GoooReadBridgeFields,[7]int32{int32(c),int32(slot-GoooReadFieldWriteStart),start,start,int32(GoooReadLocalUseDrops[c]),GoooReadBridgeCompletions[c],-1})
}
func GoooReadBridgePartStart(slot int) {
 c:=GoooReadCase;row:=&GoooReadBridgeFields[slot]
 row[2]=int32(len(GoooReadLocalUses)-GoooReadLocalUseStart)
 row[3]=row[2];row[4]=int32(GoooReadLocalUseDrops[c]);row[5]=GoooReadBridgeCompletions[c]
}
func GoooReadBridgePartEnd(slot int) {
 row:=&GoooReadBridgeFields[slot]
 row[3]=int32(len(GoooReadLocalUses)-GoooReadLocalUseStart)
 row[4]=int32(GoooReadLocalUseDrops[GoooReadCase])-row[4]
}
func GoooReadBridgeFieldEmpty(slot int) {
 row:=&GoooReadBridgeFields[slot];row[3]=row[2];row[4]=0
}
func GoooReadBridgeFieldFinish(slot int,done,part bool) {
 c:=GoooReadCase;row:=&GoooReadBridgeFields[slot]
 if !part {row[3]=int32(len(GoooReadLocalUses)-GoooReadLocalUseStart);row[4]=int32(GoooReadLocalUseDrops[c])-row[4]}
 if done {row[6]=GoooReadBridgeCompletions[c];GoooReadBridgeCompletions[c]++}
}
func GoooReadBridgeSave() {
 envelope:=map[string]any{"schema":"gooo/native-field-local-bridge/v1","locals":GoooReadBridgeLocals,"fields":GoooReadBridgeFields,
 "local_clocks":GoooReadBridgeLocalClocks,"field_clocks":GoooReadBridgeFieldClocks,"completions":GoooReadBridgeCompletions[:GoooReadCases]}
 raw,err:=json.Marshal(envelope);if err!=nil {os.Exit(4)}
 if os.WriteFile("gooo-native-field-local-bridge.json",raw,0600)!=nil {os.Exit(4)}
}
`
