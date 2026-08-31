package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagepackageexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/packageexecution"
)

func buildCases(root string) ([]languagepackageexecution.CaseEvidence, error) {
	sources, err := packageexecution.LoadDirectory(filepath.Join(root, "examples", "billing-package"))
	if err != nil {
		return nil, err
	}
	request := packageexecution.Request{PackagePath: "billing-package", Entry: "PayOrder", Sources: sources}
	positive := packageexecution.Execute(request)
	mismatch := append([]packageexecution.Source(nil), sources...)
	mismatch[0].Content = strings.Replace(mismatch[0].Content, "package billing", "package other", 1)
	if mismatch[0].Content == sources[0].Content {
		return nil, fmt.Errorf("witness: package header fixture not found")
	}
	duplicate := append(append([]packageexecution.Source(nil), sources...), packageexecution.Source{Filename: "duplicate.gooo", Content: sources[1].Content})
	return []languagepackageexecution.CaseEvidence{
		{ID: "positive-package-execution", Receipt: positive},
		{ID: "deterministic-replay", Receipt: packageexecution.Execute(request)},
		{ID: "header-mismatch-rejection", Receipt: packageexecution.Execute(packageexecution.Request{PackagePath: "billing-package", Entry: "PayOrder", Sources: mismatch})},
		{ID: "duplicate-declaration-rejection", Receipt: packageexecution.Execute(packageexecution.Request{PackagePath: "billing-package", Entry: "PayOrder", Sources: duplicate})},
		{ID: "source-count-rejection", Receipt: packageexecution.Execute(packageexecution.Request{PackagePath: "billing-package", Entry: "PayOrder", Sources: sources[:1]})},
	}, nil
}
