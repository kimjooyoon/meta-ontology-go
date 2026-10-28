package syntaxregistration

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func nativePartitionFixture(t *testing.T, directory string, names []string, filename, boundary, into string,
	support syntax.EntityFieldsSupport) (fs.FS, sourcebundle.Partition) {
	t.Helper()
	repository := os.DirFS("../../..")
	files, err := fs.Sub(repository, directory)
	if err != nil {
		t.Fatal(err)
	}
	sources, err := sourcebundle.Read(files, names)
	if err != nil {
		t.Fatal(err)
	}
	if boundary == "" {
		bundle, err := sourcebundle.LowerWithEntityFieldsSupport(sources, support)
		if err != nil {
			t.Fatal(err)
		}
		boundary = bundle.Provenance.Declarations[1].ID
	}
	partition, err := sourcebundle.PreparePartitionWithEntityFieldsSupport(sources, filename, boundary, into, support)
	if err != nil {
		t.Fatal(err)
	}
	return repository, partition
}

func TestPartitionCandidateIncludesNativeOwnersAndPhysicalSources(t *testing.T) {
	for _, item := range []struct {
		directory, filename, boundary, into string
		names                               []string
		support                             syntax.EntityFieldsSupport
		writes                              int
	}{
		{"internal/meta/executioncontinuation/policy", "source_recovery.gooo", "gooo://self-improvement/consumer/next-observation", "source_observation.gooo",
			[]string{"source.gooo", "source_recovery.gooo"}, syntax.CurrentEntityFieldsSupport(), 5},
		{"examples/entity-fields-v1", "main.gooo", "", "activity.gooo", []string{"main.gooo"}, syntax.EntityFieldsV1Support(), 4},
		{"examples/agent-tool-inventory", "main.gooo", "urn:gooo:agent-tool-inventory:inspect", "inspect.gooo",
			[]string{"main.gooo"}, syntax.EntityFieldsV5Support(), 4},
		{"examples/billing-package", "entities.gooo", "urn:gooo:billing:receipt", "receipt.gooo",
			[]string{"activity.gooo", "entities.gooo"}, syntax.CurrentEntityFieldsSupport(), 4},
	} {
		t.Run(item.directory, func(t *testing.T) {
			repository, partition := nativePartitionFixture(t, item.directory, item.names, item.filename, item.boundary, item.into, item.support)
			candidate, err := PreparePartitionRegistrationCandidate(repository, item.directory, partition)
			if err != nil {
				t.Fatal(err)
			}
			original, err := fs.ReadFile(repository, corpusPath)
			if err != nil {
				t.Fatal(err)
			}
			var baseline languagesyntax.Registry
			if err := json.Unmarshal(original, &baseline); err != nil {
				t.Fatal(err)
			}
			if len(candidate.Outputs) != item.writes || !reflect.DeepEqual(candidate.Projection.Registry.Cases, baseline.Cases) ||
				!bytes.Equal(candidate.Outputs[corpusPath], candidate.Projection.Corpus) {
				for i, prior := range baseline.Cases {
					if i < len(candidate.Projection.Registry.Cases) && !reflect.DeepEqual(prior, candidate.Projection.Registry.Cases[i]) {
						t.Logf("first case difference: before=%+v after=%+v", prior, candidate.Projection.Registry.Cases[i])
						break
					}
				}
				t.Fatal("native construction omitted an owner or changed the case population", sortedPaths(candidate.Outputs))
			}
			for _, file := range partition.Files {
				if string(candidate.Outputs[path.Join(item.directory, file.Filename)]) != file.Source {
					t.Fatal("partition source bytes changed")
				}
			}
			for _, name := range item.names {
				raw, err := fs.ReadFile(repository, path.Join(item.directory, name))
				if err != nil || !bytes.Equal(raw, candidate.Inputs[path.Join(item.directory, name)]) {
					t.Fatal("unchanged source member was not pinned", err)
				}
			}
			if err := ValidatePartitionRegistrationCandidate(repository, item.directory, partition, candidate); err != nil {
				t.Fatal(err)
			}
			if _, err := fs.Stat(repository, path.Join(item.directory, item.into)); !os.IsNotExist(err) {
				t.Fatal("preparation installed a source fragment")
			}
			assertCompiledPartitionRegistry(t, candidate)
		})
	}
}

func TestPartitionCandidateRefusesStaleInputsAndAlteredOutputs(t *testing.T) {
	const directory = "internal/meta/executioncontinuation/policy"
	repository, partition := nativePartitionFixture(t, directory, []string{"source.gooo", "source_recovery.gooo"},
		"source_recovery.gooo", "gooo://self-improvement/consumer/next-observation", "source_observation.gooo", syntax.CurrentEntityFieldsSupport())
	candidate, err := PreparePartitionRegistrationCandidate(repository, directory, partition)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{directory + "/source.gooo", syntaxRoot + "case_sources.go", corpusPath} {
		changed := append(bytes.Clone(candidate.Inputs[name]), []byte("\n ")...)
		files := sourcebundle.Overlay{FS: repository, Files: map[string][]byte{name: changed}}
		if err := ValidatePartitionRegistrationCandidate(files, directory, partition, candidate); err == nil {
			t.Fatal("stale input accepted", name)
		}
	}
	candidate.Outputs[corpusPath] = append(candidate.Outputs[corpusPath], '\n')
	if err := ValidatePartitionRegistrationCandidate(repository, directory, partition, candidate); err == nil {
		t.Fatal("altered corpus accepted")
	}
	changed := &changingPartitionNative{FS: repository, name: syntaxRoot + "case_sources.go"}
	if result, err := PreparePartitionRegistrationCandidate(changed, directory, partition); err == nil ||
		!reflect.DeepEqual(result, PartitionRegistrationCandidate{}) {
		t.Fatal("native source changed during preparation", err)
	}
}

type changingPartitionNative struct {
	fs.FS
	name  string
	reads int
}

func TestPartitionCandidatePinsAConsumerOutsideTheSelectedSplit(t *testing.T) {
	const directory = "internal/meta/executioncontinuation/policy"
	repository, partition := nativePartitionFixture(t, directory, []string{"source.gooo"}, "source.gooo",
		"gooo://self-improvement/consumer/workflow-path", "source_workflow.gooo", syntax.CurrentEntityFieldsSupport())
	candidate, err := PreparePartitionRegistrationCandidate(repository, directory, partition)
	if err != nil {
		t.Fatal(err)
	}
	name := directory + "/source_recovery.gooo"
	if len(candidate.Inputs[name]) == 0 {
		t.Fatal("unselected consumer member was not pinned")
	}
	changed := append(bytes.Clone(candidate.Inputs[name]), []byte("\n// changed origin\n")...)
	files := sourcebundle.Overlay{FS: repository, Files: map[string][]byte{name: changed}}
	if err := ValidatePartitionRegistrationCandidate(files, directory, partition, candidate); err == nil {
		t.Fatal("unselected consumer change was ignored")
	}
}

func (f *changingPartitionNative) Open(name string) (fs.File, error) {
	if name == f.name {
		f.reads++
		if f.reads == 2 {
			raw, err := fs.ReadFile(f.FS, name)
			if err != nil {
				return nil, err
			}
			return (sourcebundle.Overlay{Files: map[string][]byte{name: append(raw, []byte("\n// changed\n")...)}}).Open(name)
		}
	}
	return f.FS.Open(name)
}
