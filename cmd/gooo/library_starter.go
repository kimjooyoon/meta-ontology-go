package main

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
)

//go:embed templates/library/* templates/diagnostic/*
var libraryStarterTemplateFS embed.FS

var libraryStarterFiles = readStarterFiles("library")
var diagnosticStarterFiles = readStarterFiles("diagnostic")

func readStarterFiles(template string) map[string]string {
	directory := "templates/" + template
	entries, err := fs.ReadDir(libraryStarterTemplateFS, directory)
	if err != nil {
		panic(fmt.Sprintf("read embedded Gooo %s starter: %v", template, err))
	}
	files := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, err := libraryStarterTemplateFS.ReadFile(directory + "/" + entry.Name())
		if err != nil {
			panic(fmt.Sprintf("read embedded Gooo %s starter file %q: %v", template, entry.Name(), err))
		}
		name := strings.TrimSuffix(entry.Name(), ".template")
		files[name] = string(content)
	}
	return files
}
