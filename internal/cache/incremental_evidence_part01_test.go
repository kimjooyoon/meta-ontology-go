package cache

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"time"
)

const (
	incrementalPartCount = 10
)

var incrementalFixtureSizes = []int{10, 100, 1000, 10000}

type incrementalFixture struct {
	size  int
	ir    semantic.IR
	facts []semantic.Fact
}
type incrementalPart struct {
	index         int
	factDigests   []string
	closureIDs    []string
	closureDigest Digest
	spec          PartialSpec
	key           Key
}
type incrementalMutation struct {
	name             string
	presentationOnly bool
	mutatedFactIndex int
	dependencyPart   int
	expectedAffected func(size int) map[int]bool
}
type incrementalMeasurement struct {
	hits           int
	misses         int
	recomputations int
	started        time.Time
}
