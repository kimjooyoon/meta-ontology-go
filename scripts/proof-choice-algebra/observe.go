package main

import (
	"encoding/json"
	"os"
	"os/exec"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/proofchoicejudge"
)

func judge(source, receipt, beforePath, afterPath string) any {
	sourceData, err := os.ReadFile(source)
	if err != nil {
		return proofchoicejudge.Judge(nil)
	}
	receiptData, err := os.ReadFile(receipt)
	if err != nil {
		return proofchoicejudge.Judge(nil)
	}
	before, _ := os.ReadFile(beforePath)
	after, _ := os.ReadFile(afterPath)
	return proofchoicejudge.JudgeSource(source, sourceData, receiptData, before, after)
}

func snapshot(root string) []byte {
	if root == "" {
		return nil
	}
	command := exec.Command("git", "-C", root, "status", "--porcelain=v1", "--untracked-files=all")
	data, err := command.Output()
	if err != nil {
		return nil
	}
	return data
}

func writeSnapshot(path string, data []byte) {
	if path != "" {
		_ = os.WriteFile(path, data, 0o644)
	}
}

func decisionOf(result any) string {
	data, _ := json.Marshal(result)
	var envelope struct {
		Decision string `json:"decision"`
	}
	_ = json.Unmarshal(data, &envelope)
	return envelope.Decision
}
