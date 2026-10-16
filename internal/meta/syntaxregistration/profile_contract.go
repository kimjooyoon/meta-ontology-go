package syntaxregistration

import (
	_ "embed"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

//go:embed profile_migration.gooo
var profileMigrationSource []byte

const profileMigrationDomain = "gooo://syntax-profile-migration/"

type profileContract struct {
	source, semantic string
	pin, operation   string
	activities       [3]string
	outputs          [3]string
}

func bindProfileContract(raw []byte) (profileContract, error) {
	file, diagnostics := syntax.ParseFile("profile_migration.gooo", string(raw))
	if file == nil || diagnostics.HasErrors() {
		return profileContract{}, fmt.Errorf("PROFILE_CONTRACT_SYNTAX_INVALID")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return profileContract{}, err
	}
	if len(ir.Graph.Nodes()) != 11 {
		return profileContract{}, fmt.Errorf("PROFILE_CONTRACT_POPULATION_MISMATCH")
	}
	binding := profileContract{source: digest(raw), semantic: ir.StableHash(),
		pin: profileMigrationDomain + "pin-input", operation: profileMigrationDomain + "migrate"}
	if err := bindProfileActivity(ir.Graph, "pin-input", "request", "pinned-input", ""); err != nil {
		return profileContract{}, err
	}
	if err := bindProfileActivity(ir.Graph, "migrate", "pinned-input", "candidate", "syntax.profile-migrate:v1"); err != nil {
		return profileContract{}, err
	}
	for index, role := range [3][2]string{{"revise-corpus", "corpus"},
		{"revise-native-registry", "native-registry"}, {"retain-program", "program"}} {
		if err := bindProfileActivity(ir.Graph, role[0], "pinned-input", role[1], ""); err != nil {
			return profileContract{}, err
		}
		binding.activities[index], binding.outputs[index] = profileMigrationDomain+role[0], profileMigrationDomain+role[1]
	}
	return binding, nil
}

func bindProfileActivity(graph semantic.Graph, activity, input, output, program string) error {
	a, found := graph.Node(semantic.ID(profileMigrationDomain + activity))
	i, inputFound := graph.Node(semantic.ID(profileMigrationDomain + input))
	o, outputFound := graph.Node(semantic.ID(profileMigrationDomain + output))
	if !found || !inputFound || !outputFound || a.Kind != semantic.Activity ||
		i.Kind != semantic.Entity || o.Kind != semantic.Entity || a.ValueProgram != program ||
		!graph.HasFact(semantic.FactKey{Subject: a.ID, Predicate: semantic.Used, Object: i.ID}) ||
		!graph.HasFact(semantic.FactKey{Subject: o.ID, Predicate: semantic.WasGeneratedBy, Object: a.ID}) {
		return fmt.Errorf("PROFILE_CONTRACT_ROLE_MISMATCH: %s", profileMigrationDomain+activity)
	}
	return nil
}
