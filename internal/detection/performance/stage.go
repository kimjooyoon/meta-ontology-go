package performance

import "fmt"

// Stage names the compiler pipeline boundary being measured.
type Stage string

const (
	ParserStage    Stage = "parser"
	SemanticStage  Stage = "semantic"
	QueryStage     Stage = "query"
	GeneratorStage Stage = "generator"
	CacheStage     Stage = "cache"
)

var standardStages = [...]Stage{
	ParserStage,
	SemanticStage,
	QueryStage,
	GeneratorStage,
	CacheStage,
}

// StandardStages returns the canonical stage order used by reports and examples.
func StandardStages() []Stage {
	stages := make([]Stage, len(standardStages))
	copy(stages, standardStages[:])
	return stages
}

// Counter records work units selected by a stage adapter.
type Counter struct {
	operations uint64
}

func (c *Counter) reset() {
	c.operations = 0
}

// Add records n deterministic work units.
func (c *Counter) Add(n uint64) {
	c.operations += n
}

// Inc records one deterministic work unit.
func (c *Counter) Inc() {
	c.Add(1)
}

// Operations returns the number of work units recorded so far.
func (c Counter) Operations() uint64 {
	return c.operations
}

// Runner executes one isolated stage operation. It must not retain the counter
// pointer after returning because the harness reuses that counter.
type Runner func(*Counter) error

// StageSpec connects a stage implementation to its budget.
type StageSpec struct {
	Stage  Stage
	Run    Runner
	Budget Budget
}

func (s StageSpec) validate() error {
	if s.Stage == "" {
		return fmt.Errorf("performance stage is empty")
	}
	if s.Run == nil {
		return fmt.Errorf("performance stage %q has no runner", s.Stage)
	}
	return nil
}

func validateSpecs(specs []StageSpec) error {
	seen := make(map[Stage]struct{}, len(specs))
	for _, spec := range specs {
		if err := spec.validate(); err != nil {
			return err
		}
		if _, exists := seen[spec.Stage]; exists {
			return fmt.Errorf("performance stage %q is registered more than once", spec.Stage)
		}
		seen[spec.Stage] = struct{}{}
	}
	return nil
}
