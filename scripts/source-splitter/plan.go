package main

import (
	"fmt"
	"path"
	"path/filepath"
)

func planSource(root, subject string, limit int) (splitPlan, error) {
	var plan splitPlan
	target, err := secureSourcePath(root, subject)
	if err != nil {
		return plan, err
	}
	info, fset, file, err := parseSource(target, subject)
	if err != nil {
		return plan, err
	}
	declarations := sourceDeclarations(file)
	groups, err := partitionDeclarations(fset, file, declarations, limit)
	if err != nil {
		return plan, fmt.Errorf("%w: %s", errSplitBlocked, err)
	}
	if len(groups) < 2 {
		return plan, fmt.Errorf("%w: %s does not require declaration splitting", errSplitBlocked, subject)
	}
	plan = splitPlan{Directory: path.Dir(subject), Mode: info.Mode(), Parts: make([]splitPart, len(groups))}
	for index, group := range groups {
		partSubject, nameErr := splitPartPath(subject, index+1)
		if nameErr != nil {
			return splitPlan{}, nameErr
		}
		data, renderErr := renderPart(fset, file, group)
		if renderErr != nil {
			return splitPlan{}, renderErr
		}
		partPath := filepath.Join(filepath.Dir(target), filepath.Base(filepath.FromSlash(partSubject)))
		plan.Parts[index] = splitPart{Path: partPath, Subject: partSubject, Data: data}
	}
	return plan, nil
}
