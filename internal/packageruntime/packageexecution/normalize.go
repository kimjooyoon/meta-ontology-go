package packageexecution

import (
	"path"
	"sort"
	"strings"
)

func normalizeRequest(request Request) (Request, *Diagnostic) {
	request.PackagePath = normalizePath(request.PackagePath)
	if invalidRelativePath(request.PackagePath) {
		return Request{}, diagnostic("INPUT", "PACKAGE_PATH_INVALID", request.PackagePath, "package path must be relative and canonical")
	}
	if strings.TrimSpace(request.Entry) == "" {
		return Request{}, diagnostic("INPUT", "PACKAGE_ENTRY_INVALID", "", "entry activity is required")
	}
	if len(request.Sources) < 2 {
		return request, diagnostic("INPUT", "PACKAGE_SOURCE_COUNT_INVALID", "", "a multi-file package requires at least two .gooo files")
	}
	sources, issue := normalizeSources(request.Sources)
	if issue != nil {
		return Request{}, issue
	}
	request.Sources = sources
	return request, nil
}

func normalizeSources(values []Source) ([]Source, *Diagnostic) {
	sources := append([]Source(nil), values...)
	for index := range sources {
		sources[index].Filename = normalizePath(sources[index].Filename)
		if invalidRelativePath(sources[index].Filename) || path.Ext(sources[index].Filename) != ".gooo" {
			return nil, diagnostic("INPUT", "PACKAGE_SOURCE_PATH_INVALID", sources[index].Filename, "source path must name a relative .gooo file")
		}
	}
	sort.Slice(sources, func(left, right int) bool { return sources[left].Filename < sources[right].Filename })
	for index := 1; index < len(sources); index++ {
		if sources[index-1].Filename == sources[index].Filename {
			return nil, diagnostic("INPUT", "PACKAGE_SOURCE_DUPLICATE", sources[index].Filename, "source filename is duplicated")
		}
	}
	return sources, nil
}

func normalizePath(value string) string {
	return path.Clean(strings.ReplaceAll(strings.TrimSpace(value), "\\", "/"))
}

func invalidRelativePath(value string) bool {
	return value == "." || value == "" || path.IsAbs(value) || value == ".." || strings.HasPrefix(value, "../")
}

func diagnostic(stage, code, filename, message string) *Diagnostic {
	return &Diagnostic{Stage: stage, Code: code, Filename: filename, Message: message}
}
