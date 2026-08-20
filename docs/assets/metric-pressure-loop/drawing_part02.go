package main

import (
	"image"
	"strings"
)

func circle(img *image.Paletted, cx, cy, radius int, colorIndex uint8) {
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			dx, dy := x-cx, y-cy
			if dx*dx+dy*dy <= radius*radius {
				img.SetColorIndex(x, y, colorIndex)
			}
		}
	}
}
func diamond(img *image.Paletted, cx, cy, radius int, colorIndex uint8) {
	for offset := 0; offset <= radius; offset++ {
		line(img, cx-offset, cy-radius+offset, cx+offset, cy-radius+offset, colorIndex)
		line(img, cx-offset, cy+radius-offset, cx+offset, cy+radius-offset, colorIndex)
	}
}
func cross(img *image.Paletted, x0, y0, x1, y1 int, colorIndex uint8) {
	line(img, x0, y0, x1, y1, colorIndex)
}
func drawText(img *image.Paletted, x, y int, value string, scale int, colorIndex uint8) int {
	value = strings.ToUpper(value)
	start := x
	for i := 0; i < len(value); i++ {
		glyph, ok := glyphs[value[i]]
		if !ok {
			glyph = glyphs[' ']
		}
		for row := range 7 {
			for col := range 5 {
				if glyph[row*5+col] == '1' {
					fill(img, x+col*scale, y+row*scale, x+(col+1)*scale, y+(row+1)*scale, colorIndex)
				}
			}
		}
		x += 6 * scale
	}
	return x - start
}
func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
func float64Sin(v float64) float64 {
	if v < 0 {
		return -float64Sin(-v)
	}
	for v > 6.28318 {
		v -= 6.28318
	}
	term, sum := v, v
	for n := 3; n <= 11; n += 2 {
		term *= -v * v / float64((n-1)*n)
		sum += term
	}
	return sum
}
