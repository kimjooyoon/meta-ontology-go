package protectedregions

type sourceLine struct {
	start int
	end   int
	text  string
}
type markerEvent struct {
	kind         MarkerKind
	boundary     MarkerBoundary
	id           string
	semanticKind string
	legacy       bool
	line         int
	start        int
	end          int
}
type openMarker struct {
	event     markerEvent
	bodyStart int
}
