package main

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
)

//go:embed templates/library/*
var libraryStarterTemplateFS embed.FS

var libraryStarterFiles = readLibraryStarterFiles()

func readLibraryStarterFiles() map[string]string {
	entries, err := fs.ReadDir(libraryStarterTemplateFS, "templates/library")
	if err != nil {
		panic(fmt.Sprintf("read embedded Gooo library starter: %v", err))
	}
	files := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, err := libraryStarterTemplateFS.ReadFile("templates/library/" + entry.Name())
		if err != nil {
			panic(fmt.Sprintf("read embedded Gooo library starter file %q: %v", entry.Name(), err))
		}
		name := strings.TrimSuffix(entry.Name(), ".template")
		files[name] = string(content)
	}
	return files
}
