package main

func loopState(frame int) phaseState {
	p := float64(frame) / float64(frameCount-1)
	s := phaseState{stage: stageFloor, epoch: 1, floor: 0.28, ceiling: 0.28}
	switch {
	case p < 0.10:
		s.floor = lerp(0.28, 0.52, segment(p, 0.00, 0.10))
		s.ceiling = s.floor
	case p < 0.22:
		s.stage = stageSelect
		s.floor, s.ceiling = 0.52, 0.52
		s.selection = segment(p, 0.10, 0.22)
	case p < 0.35:
		s.stage = stageFanout
		s.floor, s.ceiling = 0.52, 0.52
		s.selection, s.attempts = 1, segment(p, 0.22, 0.35)
	case p < 0.49:
		s.stage = stageReject
		s.floor, s.ceiling = 0.52, 0.52
		s.selection, s.attempts = 1, 1
		s.reject = segment(p, 0.35, 0.49)
	case p < 0.68:
		s.stage = stageAccept
		s.floor, s.attempts, s.accepted = 0.52, 1, segment(p, 0.49, 0.68)
		s.ceiling = lerp(0.52, 0.90, s.accepted)
	case p < 0.77:
		s.stage = stageProof
		s.floor, s.ceiling, s.attempts = 0.52, 0.90, 1
		s.proof = segment(p, 0.68, 0.77)
	case p < 0.86:
		s.stage = stageRatchet
		s.floor, s.ceiling, s.attempts, s.proof = 0.52, 0.90, 1, 1
		s.sealed = segment(p, 0.77, 0.86)
	case p < 0.94:
		s.stage, s.epoch = stageEpochTwo, 2
		s.floor, s.ceiling, s.attempts, s.proof, s.sealed = 0.90, 0.90, 1, 1, 1
		s.epochTwo = segment(p, 0.86, 0.94)
		s.selection = segment(p, 0.90, 0.94)
	default:
		s.stage, s.epoch = stageNextWork, 2
		s.floor, s.ceiling, s.attempts, s.proof, s.sealed = 0.90, 0.90, 1, 1, 1
		s.epochTwo, s.selection = 1, 1
		s.nextWork = segment(p, 0.94, 1.00)
	}
	return s
}
func segment(p, start, end float64) float64 {
	if p <= start {
		return 0
	}
	if p >= end {
		return 1
	}
	return smooth((p - start) / (end - start))
}
func smooth(v float64) float64 {
	return v * v * (3 - 2*v)
}
func lerp(a, b, t float64) float64 {
	return a + (b-a)*clamp(t)
}
