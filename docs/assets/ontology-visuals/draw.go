package main

import (
	"image"
	"image/draw"
	"strings"
)

func fill(img *image.Paletted, x0, y0, x1, y1 int, c uint8) {
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 > width {
		x1 = width
	}
	if y1 > height {
		y1 = height
	}
	draw.Draw(img, image.Rect(x0, y0, x1, y1), &image.Uniform{palette[c]}, image.Point{}, draw.Src)
}
func line(img *image.Paletted, x0, y0, x1, y1 int, c uint8) {
	dx, sx := abs(x1-x0), 1
	if x0 > x1 {
		sx = -1
	}
	dy, sy := -abs(y1-y0), 1
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	for {
		if x0 >= 0 && x0 < width && y0 >= 0 && y0 < height {
			img.SetColorIndex(x0, y0, c)
		}
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}
func stroke(img *image.Paletted, x0, y0, x1, y1 int, c uint8) {
	line(img, x0, y0, x1, y0, c)
	line(img, x0, y1, x1, y1, c)
	line(img, x0, y0, x0, y1, c)
	line(img, x1, y0, x1, y1, c)
}
func box(img *image.Paletted, x0, y0, x1, y1 int, fillC, borderC uint8) {
	fill(img, x0, y0, x1, y1, fillC)
	stroke(img, x0, y0, x1, y1, borderC)
}
func circle(img *image.Paletted, cx, cy, r int, c uint8) {
	for y := cy - r; y <= cy+r; y++ {
		for x := cx - r; x <= cx+r; x++ {
			dx, dy := x-cx, y-cy
			if dx*dx+dy*dy <= r*r && x >= 0 && x < width && y >= 0 && y < height {
				img.SetColorIndex(x, y, c)
			}
		}
	}
}
func diamond(img *image.Paletted, cx, cy, r int, c uint8) {
	for d := 0; d <= r; d++ {
		line(img, cx-d, cy-r+d, cx+d, cy-r+d, c)
		line(img, cx-d, cy+r-d, cx+d, cy+r-d, c)
	}
}
func cross(img *image.Paletted, x0, y0, x1, y1 int, c uint8) {
	line(img, x0, y0, x1, y1, c)
	line(img, x1, y0, x0, y1, c)
}
func arrow(img *image.Paletted, x, y int, c uint8) {
	line(img, x, y, x-7, y-5, c)
	line(img, x, y, x-7, y+5, c)
}
func drawText(img *image.Paletted, x, y int, value string, scale int, c uint8) int {
	value = strings.ToUpper(value)
	start := x
	for i := 0; i < len(value); i++ {
		glyph := glyphs[value[i]]
		if glyph == "" {
			glyph = glyphs[' ']
		}
		for row := 0; row < 7; row++ {
			for col := 0; col < 5; col++ {
				if glyph[row*5+col] == '1' {
					fill(img, x+col*scale, y+row*scale, x+(col+1)*scale, y+(row+1)*scale, c)
				}
			}
		}
		x += 6 * scale
	}
	return x - start
}
func textLines(img *image.Paletted, x, y int, lines []string, scale int, c uint8) {
	for i, s := range lines {
		drawText(img, x, y+i*(8*scale+4), s, scale, c)
	}
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
func smooth(v float64) float64 { return v * v * (3 - 2*v) }
func progress(frame, start, end int) float64 {
	if frame <= start {
		return 0
	}
	if frame >= end {
		return 1
	}
	return smooth(float64(frame-start) / float64(end-start))
}
func activeColor(frame, start, end int, base uint8) uint8 {
	if progress(frame, start, end) > 0.2 {
		return base
	}
	return border
}

func drawNode(img *image.Paletted, x, y, w, h int, title string, lines []string, c uint8, active bool) {
	fillC := panel
	if active {
		fillC = panelRaised
	}
	box(img, x, y, x+w, y+h, fillC, activeColor(1, 0, 1, c))
	drawText(img, x+14, y+14, title, 2, c)
	textLines(img, x+14, y+45, lines, 1, textPrimary)
}
func flow(img *image.Paletted, x0, y0, x1, y1 int, c uint8, active bool) {
	line(img, x0, y0, x1, y1, c)
	if active {
		arrow(img, x1, y1, c)
	} else {
		line(img, x1-7, y1-5, x1, y1, border)
		line(img, x1-7, y1+5, x1, y1, border)
	}
}
func badge(img *image.Paletted, x, y int, label string, c uint8, shape byte) {
	if shape == 'o' {
		circle(img, x+12, y+12, 10, c)
	} else if shape == 'd' {
		diamond(img, x+12, y+12, 11, c)
	} else {
		fill(img, x, y, x+24, y+24, c)
	}
	drawText(img, x+31, y+8, label, 1, textPrimary)
}
