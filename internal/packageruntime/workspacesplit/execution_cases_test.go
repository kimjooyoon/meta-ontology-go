package workspacesplit_test

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"testing"
)

func jointCases(t *testing.T, entry string) (bodyexecution.CompositionCases, int) {
	t.Helper()
	raw, total := `{"schema":"gooo/body-composition-cases/v1","cases":[
 {"inputs":{"shared/app:Credit":2},"expected":{"shared/app:Credit":9,"shared/app:Run":6}},
 {"inputs":{"shared/app:Credit":-10},"expected":{"shared/app:Credit":-3,"shared/app:Run":-6}}
]}`, 4
	if entry == "Credit" {
		raw, total = `{"schema":"gooo/body-composition-cases/v1","cases":[
 {"inputs":{"shared/app:Credit":2},"expected":{"shared/app:Credit":9}},
 {"inputs":{"shared/app:Credit":-10},"expected":{"shared/app:Credit":-3}}
]}`, 2
	}
	suite, err := bodyexecution.DecodeCompositionCases([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	return suite, total
}
