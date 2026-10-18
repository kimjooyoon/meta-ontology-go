package syntaxregistration

import (
	"fmt"
	"go/ast"
	"slices"
	"strings"
)

const historyChunkPrefix = "historicalSyntaxTargetChunk"
const historyChunkRows = 24
const historySelector = "if target, known := historicalSyntaxTargetChunk0(version); known {\n\t\texpected[0].Target = target\n\t}"

func readHistoryChunks(source *goSource, targets map[int]int) ([]*ast.FuncDecl, error) {
	var helpers []*ast.FuncDecl
	for index := 0; index < len(source.file.Decls); index++ {
		helper, err := source.function(fmt.Sprintf("%s%d", historyChunkPrefix, index))
		if err != nil {
			return nil, err
		}
		if !historyHelperSignature(source, helper) || helper.Body == nil || len(helper.Body.List) != 2 {
			return nil, fmt.Errorf("historical syntax helper body differs")
		}
		selection, ok := helper.Body.List[0].(*ast.SwitchStmt)
		if !ok || !historySwitch(source, selection) {
			return nil, fmt.Errorf("historical syntax helper selector differs")
		}
		if err := readHistoryCases(source, selection, targets, true); err != nil {
			return nil, err
		}
		helpers = append(helpers, helper)
		terminal, err := historyContinuation(source, helper.Body.List[1], index)
		if err != nil {
			return nil, err
		}
		if terminal {
			return helpers, nil
		}
	}
	return nil, fmt.Errorf("historical syntax helper chain has no terminal")
}

func writeHistoryOrigins(source *goSource, history historyOrigins) {
	source.replace(history.selector, historySelector)
	for _, helper := range history.helpers {
		source.replace(helper, "")
	}
	versions := make([]int, 0, len(history.targets))
	for version := range history.targets {
		versions = append(versions, version)
	}
	slices.Sort(versions)
	var output strings.Builder
	for start := 0; start < len(versions); start += historyChunkRows {
		index := start / historyChunkRows
		fmt.Fprintf(&output, "\nfunc %s%d(version int) (int, bool) {\nswitch version {\n", historyChunkPrefix, index)
		end := min(start+historyChunkRows, len(versions))
		for _, version := range versions[start:end] {
			fmt.Fprintf(&output, "case %d: return %d, true\n", version, history.targets[version])
		}
		output.WriteString("}\n")
		if end < len(versions) {
			fmt.Fprintf(&output, "return %s%d(version)\n}\n", historyChunkPrefix, index+1)
		} else {
			output.WriteString("return 0, false\n}\n")
		}
	}
	source.appendAt(history.root, output.String())
}

// appendHistoryOrigin changes only the terminal helper. Previously pinned
// versions keep their source bytes and physical owners across registrations.
func appendHistoryOrigin(source *goSource, history historyOrigins, version, target int) error {
	last := history.helpers[len(history.helpers)-1]
	selection := last.Body.List[0].(*ast.SwitchStmt)
	rows := 0
	for _, statement := range selection.Body.List {
		rows += len(statement.(*ast.CaseClause).List)
	}
	if rows > historyChunkRows {
		return fmt.Errorf("historical syntax helper exceeds its bounded capacity")
	}
	if rows < historyChunkRows {
		source.insert(selection.Body.Rbrace, fmt.Sprintf("case %d: return %d, true\n", version, target))
		return nil
	}
	index := len(history.helpers)
	source.replace(last.Body.List[1], fmt.Sprintf("return %s%d(version)", historyChunkPrefix, index))
	source.appendAt(last, fmt.Sprintf("\nfunc %s%d(version int) (int, bool) {\nswitch version {\ncase %d: return %d, true\n}\nreturn 0, false\n}\n",
		historyChunkPrefix, index, version, target))
	return nil
}
