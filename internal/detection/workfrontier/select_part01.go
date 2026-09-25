package workfrontier

type pathClass uint8

const (
	pathReady pathClass = iota
	pathUnknown
	pathBlocked
	pathShortfall
)

type frontierIndexes struct {
	pressures map[string]struct{}
	states    map[string]string
	invalid   bool
}
