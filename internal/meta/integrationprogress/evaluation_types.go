package integrationprogress

import "time"

type runTiming struct {
	Created time.Time
	Started time.Time
	Updated time.Time
}

type pullEvaluation struct {
	Cells                     []Cell
	QueueSeconds              int64
	ExecutionSeconds          int64
	EvidenceLatencySeconds    int64
	MergeAfterEvidenceSeconds int64
	TimingSample              bool
	EvidenceLatencySample     bool
	MergeDelaySample          bool
}
