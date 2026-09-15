package main

import (
	"image"
)

const (
	stageFloor = iota
	stageSelect
	stageFanout
	stageReject
	stageAccept
	stageProof
	stageRatchet
	stageEpochTwo
	stageNextWork
)

type phaseState struct {
	stage     int
	epoch     int
	floor     float64
	ceiling   float64
	selection float64
	attempts  float64
	reject    float64
	accepted  float64
	proof     float64
	sealed    float64
	epochTwo  float64
	nextWork  float64
}

func renderFrame(frame int) *image.Paletted {
	img := image.NewPaletted(image.Rect(0, 0, width, height), palette)
	fill(img, 0, 0, width, height, background)
	state := loopState(frame)

	drawHeader(img, state)
	drawFlowArrows(img, state)
	drawFloor(img, state)
	drawSelector(img, state)
	drawAgents(img, state)
	drawProvenance(img, state)
	drawCI(img, state)
	return img
}
