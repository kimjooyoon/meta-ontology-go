package main

import (
	"image"
	"image/draw"
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
