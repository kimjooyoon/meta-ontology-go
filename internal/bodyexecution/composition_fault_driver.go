package bodyexecution

import (
	"fmt"
	"go/format"
	"strconv"
	"strings"
)

func arithmeticCompositionDriver(prior Composition) (string, error) {
	prefix := arithmeticObservationPrefix(prior.Source)
	graph := compositionGraph{plan: prior.Plan, count: len(prior.Plan.Activities)}
	if graph.count < 1 || graph.count > compositionLimit {
		return "", fmt.Errorf("arithmetic observation graph exceeds its bound")
	}
	copy(graph.nodes[:], prior.Plan.Activities)
	roots := 0
	for _, node := range graph.nodes[:graph.count] {
		for _, input := range node.inputSlots() {
			if input.From < 0 {
				roots++
			}
		}
	}
	var source strings.Builder
	source.WriteString("package main\nimport(\"encoding/json\";\"os\")\nfunc main(){\n")
	fmt.Fprintf(&source, "var in [][%d]json.RawMessage;if json.NewDecoder(os.Stdin).Decode(&in)!=nil||len(in)>128{os.Exit(2)}\n", roots)
	fmt.Fprintf(&source, "out:=make([][%d]json.RawMessage,len(in));faults:=make([][%d]*%sFault,len(in));blocked:=make([][%d][]string,len(in))\n", graph.count, graph.count, prefix, graph.count)
	fmt.Fprintf(&source, "for c,row:=range in{%sCase=c;%sCount=0\n", prefix, prefix)
	root := 0
	for i, node := range graph.nodes[:graph.count] {
		root = writeArithmeticDriverStep(&source, graph, node, i, root, prefix)
	}
	fmt.Fprintf(&source, `};if json.NewEncoder(os.Stdout).Encode(%sObservation{"gooo/arithmetic-outcome-observation/v1",out,faults,blocked,%sCalls})!=nil{os.Exit(3)}}`, prefix, prefix)
	source.WriteString(namedArithmeticObserver(prefix))
	raw, err := format.Source([]byte(source.String()))
	return string(raw), err
}

func writeArithmeticDriverStep(source *strings.Builder, graph compositionGraph, node CompositionActivity, index, root int, prefix string) int {
	arguments := make([]string, 0, len(node.inputSlots()))
	seen := map[int]bool{}
	for _, slot := range node.inputSlots() {
		input := fmt.Sprintf("v%d", slot.From)
		if slot.From < 0 {
			input = fmt.Sprintf("input%d", root)
			fmt.Fprintf(source, "var %s %s;if json.Unmarshal(row[%d],&%s)!=nil{os.Exit(2)}\n", input, graph.valueGoType(slot.Type), root, input)
			root++
		} else if !seen[slot.From] {
			fmt.Fprintf(source, "if faults[c][%d]!=nil||len(blocked[c][%d])!=0{blocked[c][%d]=append(blocked[c][%d],%q)}\n", slot.From, slot.From, index, index, graph.nodes[slot.From].ID)
			seen[slot.From] = true
		}
		arguments = append(arguments, input)
	}
	output := graph.valueGoType(node.OutputType)
	fmt.Fprintf(source, "var v%d %s;if len(blocked[c][%d])==0{\n", index, output, index)
	fmt.Fprintf(source, "v%d,faults[c][%d]=%sInvoke(func()%s{return GoooComposedActivity%d(%s)})\n", index, index, prefix, output, index, strings.Join(arguments, ","))
	fmt.Fprintf(source, "if faults[c][%d]==nil{var err error;out[c][%d],err=json.Marshal(v%d);if err!=nil{os.Exit(3)}}}\n", index, index, index)
	return root
}

func arithmeticObservationPrefix(source string) string {
	prefix := "GoooObservedArithmetic"
	for suffix := 0; strings.Contains(source, prefix); suffix++ {
		prefix = "GoooObservedArithmetic" + strconv.Itoa(suffix)
	}
	return prefix
}

func namedArithmeticObserver(prefix string) string {
	return strings.NewReplacer(
		"GoooArithmeticObservation", prefix+"Observation", "GoooArithmeticFault", prefix+"Fault",
		"GoooObserveQuotient", prefix+"Quotient", "GoooObserveRemainder", prefix+"Remainder",
		"GoooObserveArithmetic", prefix+"Invoke", "GoooObservedComposition", prefix+"Composition",
		"GoooObservedCall", prefix+"Call", "GoooObserveCalledInputs", prefix+"CalledInputs",
		"goooObservationCase", prefix+"Case", "goooObservationCount", prefix+"Count", "goooObservationCalls", prefix+"Calls",
	).Replace(calledInputObserver + arithmeticObserver)
}

const arithmeticObserver = `
type GoooArithmeticObservation struct {
    Schema string ` + "`json:\"schema\"`" + `
    Outputs any ` + "`json:\"outputs\"`" + `
    Faults any ` + "`json:\"faults\"`" + `
    Blocked any ` + "`json:\"blocked\"`" + `
    Calls []GoooObservedCall ` + "`json:\"calls\"`" + `
}
type GoooArithmeticFault struct {
    Site int ` + "`json:\"site\"`" + `
    Left int64 ` + "`json:\"left\"`" + `
    Right int64 ` + "`json:\"right\"`" + `
}
func GoooObserveQuotient(left,right int64,site int) int64 {
    if right==0 { panic(GoooArithmeticFault{site,left,right}) }
    return left/right
}
func GoooObserveRemainder(left,right int64,site int) int64 {
    if right==0 { panic(GoooArithmeticFault{site,left,right}) }
    return left%right
}
func GoooObserveArithmetic[T any](call func() T)(value T,fault *GoooArithmeticFault) {
    defer func(){
        if problem:=recover();problem!=nil {
            if arithmetic,ok:=problem.(GoooArithmeticFault);ok { fault=&arithmetic } else { panic(problem) }
        }
    }()
    value=call()
    return
}
`
