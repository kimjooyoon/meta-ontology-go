package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type publicationName struct {
	requested, exact, folded, sole string
	foldedCount, sameCount         int
}

func (name *publicationName) observe(entry string) {
	if entry == name.requested {
		name.exact = entry
	}
	name.sole, name.sameCount = entry, name.sameCount+1
	if strings.EqualFold(entry, name.requested) {
		name.folded, name.foldedCount = entry, name.foldedCount+1
	}
}

func (name publicationName) resolve() (string, error) {
	switch {
	case name.exact != "":
		return name.exact, nil
	case name.foldedCount == 1:
		return name.folded, nil
	case name.sameCount == 1:
		return name.sole, nil
	default:
		return "", fmt.Errorf("source directory entry %q is ambiguous", name.requested)
	}
}

// Only two name observations and one bounded directory batch remain in memory.
func comparePublicationNames(parent, left, right string, source os.FileInfo) (bool, error) {
	directory, err := os.Open(parent)
	if err != nil {
		return false, err
	}
	defer directory.Close()
	names := [2]publicationName{{requested: left}, {requested: right}}
	for {
		entries, readErr := directory.ReadDir(64)
		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				return false, err
			}
			if os.SameFile(info, source) {
				names[0].observe(entry.Name())
				names[1].observe(entry.Name())
			}
		}
		if readErr == io.EOF || (names[0].exact != "" && names[1].exact != "") {
			break
		}
		if readErr != nil {
			return false, readErr
		}
	}
	a, err := names[0].resolve()
	if err != nil {
		return false, err
	}
	b, err := names[1].resolve()
	return a == b, err
}
