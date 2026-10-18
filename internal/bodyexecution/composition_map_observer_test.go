package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// This is a native observer-kernel test, not a Gooo generation or conformance
// result. The source-level map expression remains a separate unfinished boundary.
func TestCollectionMapObservationKernel(t *testing.T) {
	dir := t.TempDir()
	source := "package main\nimport (\"encoding/json\";\"os\")\n" + withCollectionMapObserver(constructionReadObserver) + collectionMapObserverHarness
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, nativeTool(), "run", "main.go")
	command.Dir = dir
	command.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOWORK=off", "GOPROXY=off")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("map observation kernel: %v: %s", err, output)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "gooo-native-read-observation.json"))
	if err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Constructions []struct {
			Case      int                              `json:"case_index"`
			Completed bool                             `json:"completed"`
			Elements  []CompositionConstructionElement `json:"elements"`
		} `json:"constructions"`
		Dropped []int `json:"construction_dropped"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatal(err)
	}
	if len(envelope.Constructions) != 5 || len(envelope.Dropped) != 7 || envelope.Dropped[5] != 1 {
		t.Fatal("construction retention differs", string(raw))
	}
	for i, expected := range []int{2, 0, 2, 3, 256} {
		row := envelope.Constructions[i]
		if row.Case != i || len(row.Elements) != expected || row.Completed != (i != 2) {
			t.Fatal("wrong case, attempted prefix or completion", i, row)
		}
		for j, element := range row.Elements {
			if element.ReadStart != j || element.ReadEnd != j+1 || element.ReadsDropped != 0 ||
				element.Completed != !(i == 2 && j == 1) {
				t.Fatal("mapper must consume exactly its own input read in order", i, j, element)
			}
		}
	}
}

const collectionMapObserverHarness = `
func require(ok bool) {if !ok {panic("map observer contract")}}
func main() {
 GoooReadReset(0,7)
 input:=[]int64{5,5,8}
 GoooReadRegister(input,0)
 output:=GoooReadMap(0,0,input[1:],func(n int64) int64 {return n*2})
 require(len(output)==2 && output[0]==10 && output[1]==16 && input[1]==5)
 require(GoooReadEvents[0].Origin.Index==1 && GoooReadEvents[1].Origin.Index==2)
 require(GoooReadRead(1,output,0)==10 && GoooReadRead(1,output,1)==16)
 require(GoooReadEvents[2].Construction.Instance==0 && GoooReadEvents[2].Construction.Index==0)
 require(GoooReadEvents[3].Construction.Instance==0 && GoooReadEvents[3].Construction.Index==1)

 GoooReadReset(1,7)
 calls:=0
 empty:=GoooReadMap(0,0,[]int64{},func(n int64) int64 {calls++;return n})
 require(empty!=nil && len(empty)==0 && calls==0)

 GoooReadReset(2,7)
 GoooReadRegister(input,0)
 panicked:=false
 func(){defer func(){panicked=recover()!=nil}()
  GoooReadMap(0,0,input,func(n int64) int64 {calls++;if calls==2 {panic("mapper fault")};return n})
 }()
 require(panicked && calls==2)

 GoooReadReset(3,7)
 flags:=[]bool{false,true,false}
 GoooReadRegister(flags,0)
 labels:=GoooReadMap(0,0,flags,func(b bool) string {if b {return "한글"};return "off"})
 require(len(labels)==3 && labels[0]=="off" && labels[1]=="한글" && labels[2]=="off")

 GoooReadReset(4,7)
 maximum:=make([]int64,256)
 for i:=range maximum {maximum[i]=int64(i)}
 GoooReadRegister(maximum,0)
 mapped:=GoooReadMap(0,0,maximum,func(n int64) int64 {return n+1})
 require(len(mapped)==256 && mapped[0]==1 && mapped[255]==256 && maximum[255]==255)

 GoooReadReset(5,7)
 GoooReadRegister(input,0)
 GoooReadReserved=65536
 dropped:=GoooReadMap(0,0,input,func(n int64) int64 {return n+1})
 require(len(dropped)==3 && dropped[1]==6 && GoooReadConstructionDropped[5]==1)
 require(GoooReadEvents[len(GoooReadEvents)-1].Origin.Index==2)

 GoooReadReset(6,7)
 calls=0;panicked=false
 func(){defer func(){panicked=recover()!=nil}()
  GoooReadMap(0,0,make([]int64,257),func(n int64) int64 {calls++;return n})
 }()
 require(panicked && calls==0 && len(GoooReadConstructions)==5)
 GoooReadSave(true)
}
`
