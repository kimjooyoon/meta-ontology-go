package syntaxregistration

import (
	"fmt"
	"go/ast"
	"slices"
	"strings"
)

type historyOrigins struct {
	root     *ast.FuncDecl
	selector ast.Stmt
	helpers  []*ast.FuncDecl
	targets  map[int]int
}

// Freeze the prior source-selected target before the next corpus grows.
// Physical helper partitioning never changes a historical observation.
func pinPreviousSyntaxTarget(source *goSource, version, capability int) error {
	history, err := readHistoryOrigins(source)
	if err != nil {
		return err
	}
	if target, exists := history.targets[version]; exists && target != capability {
		return fmt.Errorf("historical syntax target differs from its pinned source corpus")
	}
	history.targets[version] = capability
	writeHistoryOrigins(source, history)
	return nil
}

func readHistoryOrigins(source *goSource) (historyOrigins, error) {
	root, err := source.function("expectedBoundarySpecsForVersion")
	if err != nil {
		return historyOrigins{}, err
	}
	history := historyOrigins{root: root, targets: map[int]int{}}
	for _, statement := range root.Body.List {
		if selection, ok := statement.(*ast.SwitchStmt); ok && historySwitch(source, selection) {
			if history.selector != nil {
				return history, fmt.Errorf("historical syntax target switch is ambiguous")
			}
			history.selector = selection
			err = readHistoryCases(source, selection, history.targets, false)
		} else if isHistorySelector(source, statement) {
			if history.selector != nil {
				return history, fmt.Errorf("historical syntax selector is ambiguous")
			}
			history.selector = statement
			history.helpers, err = readHistoryChunks(source, history.targets)
		}
		if err != nil {
			return history, err
		}
	}
	if history.selector == nil {
		return history, fmt.Errorf("historical syntax target selector is missing")
	}
	for _, declaration := range source.file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && strings.HasPrefix(function.Name.Name, historyChunkPrefix) && !historyOwnsHelper(history, function) {
			return history, fmt.Errorf("historical syntax target has an orphan helper")
		}
	}
	return history, nil
}

func historyOwnsHelper(history historyOrigins, function *ast.FuncDecl) bool {
	return slices.Contains(history.helpers, function)
}

func historySwitch(source *goSource, selection *ast.SwitchStmt) bool {
	return selection.Init == nil && selection.Tag != nil && source.text(selection.Tag) == "version"
}
