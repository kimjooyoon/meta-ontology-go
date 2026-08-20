package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
)

func printSourceMetrics(root, storageRoot string) error {
	report, err := linecaps.AnalyzeProjectedLineMetrics(root, storageRoot)
	if err != nil {
		return err
	}
	total := report.Total()
	fmt.Printf("source metrics: total_files=%d total_dirs=%d go_lines=%d gooo_lines=%d\n", total.RecursiveFiles, total.RecursiveFolders, total.GoLines, total.GoooLines)
	fmt.Printf("source language totals: go_files=%d gooo_files=%d\n", total.GoFiles, total.GoooFiles)
	fmt.Printf("source metrics detail:\n%s", report.Text())
	return nil
}
