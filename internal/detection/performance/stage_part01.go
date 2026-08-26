package performance

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

// Valid reports whether s is one of the current compiler pipeline boundaries.
func (s Stage) Valid() bool {
	for _, standard := range standardStages {
		if s == standard {
			return true
		}
	}
	return false
}
func stageOrder(s Stage) int {
	for i, standard := range standardStages {
		if s == standard {
			return i
		}
	}
	return len(standardStages)
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
