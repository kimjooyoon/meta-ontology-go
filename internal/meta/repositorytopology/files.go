package repositorytopology

import (
	"bytes"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func (s *inspection) inspectFiles(root string) {
	seen := map[string]bool{}
	for _, file := range s.source.Files {
		exact := validRelativePath(file.Path) && file.Lines >= 0
		if seen[file.Path] {
			s.duplicates++
			exact = false
		}
		seen[file.Path] = true
		if file.Language != "go" && file.Language != "gooo" && file.Language != "other" {
			s.lowerResolution = true
			exact = false
		}
		actual := 0
		if file.Language == "go" || file.Language == "gooo" {
			content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file.Path)))
			if err != nil {
				exact = false
			} else {
				actual = physicalLines(content)
				exact = exact && actual == file.Lines
			}
		}
		s.actualLines[file.Path] = actual
		s.addLanguage(file.Language, actual)
		if exact {
			s.fileRowsExact++
		}
	}
}

func validRelativePath(value string) bool {
	return value != "" && value != "." && !strings.HasPrefix(value, "/") &&
		!strings.Contains(value, "\\") && path.Clean(value) == value && value != ".." &&
		!strings.HasPrefix(value, "../")
}

func physicalLines(content []byte) int {
	if len(content) == 0 {
		return 0
	}
	lines := bytes.Count(content, []byte{'\n'})
	if content[len(content)-1] != '\n' {
		lines++
	}
	return lines
}

func (s *inspection) addLanguage(language string, lines int) {
	if language == "go" {
		s.goFiles++
		s.goLines += lines
	}
	if language == "gooo" {
		s.goooFiles++
		s.goooLines += lines
	}
}
