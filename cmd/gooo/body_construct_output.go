package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func writeJointOutput(directory string, source []byte, output bodyConstructOutput) error {
	if err := os.Mkdir(directory, 0755); err != nil {
		return err
	}
	files := map[string][]byte{
		"original.gooo": source, "selected.gooo": []byte(output.Construction.SelectedSource),
		"generated.go": []byte(output.Construction.Selected.Source), "main.go": []byte(output.Construction.Selected.Driver),
		"go.mod": []byte("module gooo.observed.composition\n\ngo 1.27.1\n"),
	}
	for name, value := range map[string]any{"construction.json": output.Construction, "evaluation.json": output.Evaluation,
		"composition.json": output.Construction.Selected, "construction-cases.json": output.Construction.ConstructionCases} {
		raw, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return err
		}
		files[name] = append(raw, '\n')
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(directory, name), content, 0644); err != nil {
			return err
		}
	}
	return nil
}
