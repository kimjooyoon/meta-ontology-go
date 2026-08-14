package main

import (
	"image"
	"image/draw"
	"strings"
)

func box(img *image.Paletted, x0, y0, x1, y1 int, fillColor, borderColor uint8) {
	fill(img, x0, y0, x1, y1, fillColor)
	stroke(img, x0, y0, x1, y1, borderColor)
}

func fill(img *image.Paletted, x0, y0, x1, y1 int, colorIndex uint8) {
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
	draw.Draw(img, image.Rect(x0, y0, x1, y1), &image.Uniform{palette[colorIndex]}, image.Point{}, draw.Src)
}

func stroke(img *image.Paletted, x0, y0, x1, y1 int, colorIndex uint8) {
	line(img, x0, y0, x1, y0, colorIndex)
	line(img, x0, y1, x1, y1, colorIndex)
	line(img, x0, y0, x0, y1, colorIndex)
	line(img, x1, y0, x1, y1, colorIndex)
}

func line(img *image.Paletted, x0, y0, x1, y1 int, colorIndex uint8) {
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
			img.SetColorIndex(x0, y0, colorIndex)
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

func arrow(img *image.Paletted, x, y int, colorIndex uint8) {
	line(img, x, y, x-7, y-5, colorIndex)
	line(img, x, y, x-7, y+5, colorIndex)
}

func arrowLeft(img *image.Paletted, x, y int, colorIndex uint8) {
	line(img, x, y, x+7, y-5, colorIndex)
	line(img, x, y, x+7, y+5, colorIndex)
}

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
		for row := 0; row < 7; row++ {
			for col := 0; col < 5; col++ {
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
