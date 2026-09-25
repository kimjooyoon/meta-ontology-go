package main

import (
	"image"
)

func drawCrossLegend(img *image.Paletted, x, y int, colorIndex uint8, label string) {
	cross(img, x+2, y+2, x+12, y+12, colorIndex)
	cross(img, x+12, y+2, x+2, y+12, colorIndex)
	drawText(img, x+18, y+2, label, 1, textMuted)
}
func drawDiamondLegend(img *image.Paletted, x, y int, colorIndex uint8, label string) {
	diamond(img, x+7, y+7, 7, colorIndex)
	drawText(img, x+18, y+2, label, 1, textMuted)
}
