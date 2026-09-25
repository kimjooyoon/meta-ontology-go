package main

import (
	"image/color"
)

const (
	background uint8 = iota
	panel
	panelRaised
	border
	textPrimary
	textMuted
	cyan
	teal
	violet
	amber
	coral
	green
	red
	unknown
	white
	blue
	dark
	lime
)

var palette = color.Palette{
	color.RGBA{R: 11, G: 19, B: 36, A: 255},
	color.RGBA{R: 17, G: 29, B: 53, A: 255},
	color.RGBA{R: 22, G: 38, B: 68, A: 255},
	color.RGBA{R: 46, G: 69, B: 108, A: 255},
	color.RGBA{R: 239, G: 246, B: 255, A: 255},
	color.RGBA{R: 155, G: 176, B: 207, A: 255},
	color.RGBA{R: 56, G: 189, B: 248, A: 255},
	color.RGBA{R: 45, G: 212, B: 191, A: 255},
	color.RGBA{R: 167, G: 139, B: 250, A: 255},
	color.RGBA{R: 251, G: 191, B: 36, A: 255},
	color.RGBA{R: 251, G: 113, B: 133, A: 255},
	color.RGBA{R: 74, G: 222, B: 128, A: 255},
	color.RGBA{R: 248, G: 113, B: 113, A: 255},
	color.RGBA{R: 102, G: 112, B: 133, A: 255},
	color.RGBA{R: 255, G: 255, B: 255, A: 255},
	color.RGBA{R: 96, G: 165, B: 250, A: 255},
	color.RGBA{R: 10, G: 16, B: 32, A: 255},
	color.RGBA{R: 190, G: 242, B: 100, A: 255},
}
