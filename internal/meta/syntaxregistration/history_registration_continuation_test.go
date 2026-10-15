package syntaxregistration

import (
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

func TestPartitionedHistoryContinuesNineRegistrationRoles(t *testing.T) {
	data, request := fixture(t)
	delete(data, request.Case.Path)
	request.Case.ID = "history-partition-continuation-20261015"
	request.Case.Path = "examples/history-partition-continuation-20261015/main.gooo"
	data[request.Case.Path] = &fstest.MapFile{Data: []byte(strings.ReplaceAll(fixtureSource, "registrationfixture", "historycontinuation"))}
	pin(t, data, &request)
	initial := digestValue(data)
	_, _, capability, err := corpusTotals(data[corpusPath].Data)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := Compile(data, request)
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := plan.Generate(data)
	if err != nil {
		t.Fatal(err)
	}
	if err := plan.ValidateCandidate(data, candidate); err != nil {
		t.Fatal(err)
	}
	if len(candidate.Artifacts) != 9 || candidate.ApplyAuthorized || candidate.RepositoryWrites != 0 || digestValue(data) != initial {
		t.Fatal("proposal changed authority, artifact roles or supplied source")
	}
	before := historyContinuationSource(t, data)
	prior, err := readHistoryOrigins(before)
	if err != nil {
		t.Fatal(err)
	}
	for _, member := range candidate.Members {
		data[member.Path] = &fstest.MapFile{Data: member.Content}
	}
	after := historyContinuationSource(t, data)
	next, err := readHistoryOrigins(after)
	if err != nil {
		t.Fatal(err)
	}
	if next.targets[request.BaseVersion] != capability {
		t.Fatal("current source target was not retained")
	}
	delete(next.targets, request.BaseVersion)
	if !reflect.DeepEqual(prior.targets, next.targets) {
		t.Fatal("prior declared history changed")
	}
	t.Log("nine source-owned registration roles; one new hypothetical case; all 27 prior version targets preserved; native/model executions 0")
}

func historyContinuationSource(t *testing.T, data fstest.MapFS) *goSource {
	t.Helper()
	inputs := map[string][]byte{}
	for name, file := range data {
		inputs[name] = file.Data
	}
	source, err := parseSourceUnits(inputs, closureRoot, "verticalsliceclosureshadow", false)
	if err != nil {
		t.Fatal(err)
	}
	return source
}
