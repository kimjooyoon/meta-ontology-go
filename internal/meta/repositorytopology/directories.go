package repositorytopology

import (
	"path"
	"strings"
)

func (s *inspection) inspectDirectories() {
	seen := map[string]bool{}
	for _, directory := range s.source.Directories {
		exact := validDirectoryPath(directory.Path)
		if seen[directory.Path] {
			s.duplicates++
			exact = false
		}
		seen[directory.Path] = true
		expected := s.expectedDirectory(directory.Path)
		exact = exact && directory == expected
		if exact {
			s.directoryRowsExact++
		}
	}
}

func validDirectoryPath(value string) bool {
	return value == "." || validRelativePath(value)
}

func (s *inspection) expectedDirectory(base string) DirectoryMetric {
	kind := "DIRECTORY"
	if base == "." {
		kind = "PROJECT_ROOT"
	}
	want := DirectoryMetric{Path: base, SubjectKind: kind}
	for _, candidate := range s.source.Directories {
		if candidate.Path != "." && path.Dir(candidate.Path) == base {
			want.DirectFolders++
		}
		if descendant(base, candidate.Path) {
			want.RecursiveFolders++
		}
	}
	for _, file := range s.source.Files {
		if path.Dir(file.Path) == base {
			want.DirectFiles++
		}
		if !containsFile(base, file.Path) {
			continue
		}
		want.RecursiveFiles++
		if file.Language == "go" {
			want.GoFiles++
			want.GoLines += s.actualLines[file.Path]
		}
		if file.Language == "gooo" {
			want.GoooFiles++
			want.GoooLines += s.actualLines[file.Path]
		}
	}
	return want
}

func descendant(base, candidate string) bool {
	if candidate == "." || candidate == base {
		return false
	}
	return base == "." || strings.HasPrefix(candidate, base+"/")
}

func containsFile(base, file string) bool {
	return base == "." || strings.HasPrefix(file, base+"/")
}
