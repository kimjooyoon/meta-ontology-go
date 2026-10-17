package verify

import (
	"regexp"
	"testing"
)

// Count declared install lines without depending on an argument's quote style.
// Other source/identity checks keep ownership of the selected tool and revision.
func workDomainToolPreparationCount124073(workflow string) int {
	pattern := `(?m)^[ \t]*go[ \t]+install[ \t]+["']?github\.com/kimjooyoon/gooo-ci-domain/`
	return len(regexp.MustCompile(pattern).FindAllStringIndex(workflow, -1))
}

func TestWorkDomainInstallCountPreservesQuotedArguments(t *testing.T) {
	module := "github.com/kimjooyoon/gooo-ci-domain/cmd/gooo-ci-domain@revision"
	cases := []struct {
		name, input string
		want        int
	}{
		{"unquoted", "go install " + module + "\n", 1},
		{"double-quoted", "  go install \"" + module + "\"\n", 1},
		{"single-quoted", "\tgo install '" + module + "'\n", 1},
		{"two-installations", "go install " + module + "\ngo install \"" + module + "\"\n", 2},
		{"other-operation", "go build " + module + "\n", 0},
		{"other-module", "go install github.com/another/module/cmd/tool@revision\n", 0},
		{"comment", "# go install " + module + "\n", 0},
		{"quoted-text", "echo 'go install " + module + "'\n", 0},
		{"adjacent-word", "go installer " + module + "\n", 0},
	}
	for _, row := range cases {
		t.Run(row.name, func(t *testing.T) {
			if got := workDomainToolPreparationCount124073(row.input); got != row.want {
				t.Fatal("declared installation population changed", got, row.want)
			}
		})
	}
}
