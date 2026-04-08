package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"

	xdraw "golang.org/x/image/draw"
)

func main() {
	root, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	assetsDir := filepath.Join(root, "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		panic(err)
	}

	large := image.NewNRGBA(image.Rect(0, 0, 1024, 1024))
	drawBackground(large)
	drawDocuments(large)

	final := image.NewNRGBA(image.Rect(0, 0, 256, 256))
	xdraw.CatmullRom.Scale(final, final.Bounds(), large, large.Bounds(), xdraw.Over, nil)

	if err := writePNG(filepath.Join(assetsDir, "icon.png"), final); err != nil {
		panic(err)
	}
	if err := writeICO(filepath.Join(assetsDir, "icon.ico"), final); err != nil {
		panic(err)
	}
}

func drawBackground(img *image.NRGBA) {
	top := color.NRGBA{R: 0x9E, G: 0xE3, B: 0xFF, A: 0xFF}
	bottom := color.NRGBA{R: 0x5D, G: 0x9E, B: 0xF6, A: 0xFF}
	fillRoundedGradient(img, image.Rect(96, 96, 928, 928), 180, top, bottom)
	fillRoundedRect(img, image.Rect(124, 124, 900, 900), 156, color.NRGBA{R: 0x1D, G: 0x35, B: 0x57, A: 0x25})
}

func drawDocuments(img *image.NRGBA) {
	shadow := color.NRGBA{R: 0x2A, G: 0x4E, B: 0x79, A: 0x30}
	backSheet := color.NRGBA{R: 0xD9, G: 0xEC, B: 0xFF, A: 0xFF}
	frontSheet := color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	fold := color.NRGBA{R: 0xD5, G: 0xE6, B: 0xF7, A: 0xFF}
	ink := color.NRGBA{R: 0xB5, G: 0xC8, B: 0xDE, A: 0xFF}
	split := color.NRGBA{R: 0x2E, G: 0xB7, B: 0xD8, A: 0xFF}
	face := color.NRGBA{R: 0x26, G: 0x3D, B: 0x5B, A: 0xFF}
	blush := color.NRGBA{R: 0xFF, G: 0xC0, B: 0xD5, A: 0xFF}

	fillRoundedRect(img, image.Rect(252, 250, 586, 746), 44, shadow)
	fillRoundedRect(img, image.Rect(230, 226, 564, 722), 44, backSheet)
	fillTriangle(img, image.Point{486, 226}, image.Point{564, 226}, image.Point{564, 304}, fold)

	fillRoundedRect(img, image.Rect(340, 292, 772, 832), 52, shadow)
	fillRoundedRect(img, image.Rect(312, 264, 744, 804), 52, frontSheet)
	fillTriangle(img, image.Point{650, 264}, image.Point{744, 264}, image.Point{744, 358}, fold)

	for y := 352; y <= 640; y += 72 {
		fillRoundedRect(img, image.Rect(514, y, 542, y+42), 12, split)
	}
	fillRoundedRect(img, image.Rect(404, 366, 496, 390), 12, ink)
	fillRoundedRect(img, image.Rect(560, 366, 650, 390), 12, ink)
	fillRoundedRect(img, image.Rect(404, 438, 478, 462), 12, ink)
	fillRoundedRect(img, image.Rect(560, 438, 658, 462), 12, ink)

	fillCircle(img, image.Point{438, 560}, 18, face)
	fillCircle(img, image.Point{614, 560}, 18, face)
	fillCircle(img, image.Point{390, 604}, 24, blush)
	fillCircle(img, image.Point{662, 604}, 24, blush)
	fillRoundedRect(img, image.Rect(484, 618, 570, 640), 10, face)
	fillRoundedRect(img, image.Rect(504, 606, 550, 620), 8, frontSheet)

	fillCircle(img, image.Point{710, 430}, 10, split)
	fillCircle(img, image.Point{742, 470}, 8, split)
}

func fillRoundedGradient(img *image.NRGBA, rect image.Rectangle, radius int, top, bottom color.NRGBA) {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		t := float64(y-rect.Min.Y) / float64(rect.Dy()-1)
		c := lerp(top, bottom, t)
		for x := rect.Min.X; x < rect.Max.X; x++ {
			if insideRoundedRect(x, y, rect, radius) {
				img.SetNRGBA(x, y, c)
			}
		}
	}
}

func fillRoundedRect(img *image.NRGBA, rect image.Rectangle, radius int, c color.NRGBA) {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			if insideRoundedRect(x, y, rect, radius) {
				img.SetNRGBA(x, y, c)
			}
		}
	}
}

func fillCircle(img *image.NRGBA, center image.Point, radius int, c color.NRGBA) {
	r2 := radius * radius
	for y := center.Y - radius; y <= center.Y+radius; y++ {
		for x := center.X - radius; x <= center.X+radius; x++ {
			dx := x - center.X
			dy := y - center.Y
			if dx*dx+dy*dy <= r2 {
				if image.Pt(x, y).In(img.Bounds()) {
					img.SetNRGBA(x, y, c)
				}
			}
		}
	}
}

func fillTriangle(img *image.NRGBA, a, b, c image.Point, col color.NRGBA) {
	minX := min3(a.X, b.X, c.X)
	maxX := max3(a.X, b.X, c.X)
	minY := min3(a.Y, b.Y, c.Y)
	maxY := max3(a.Y, b.Y, c.Y)
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if pointInTriangle(float64(x), float64(y), a, b, c) {
				img.SetNRGBA(x, y, col)
			}
		}
	}
}

func insideRoundedRect(x, y int, rect image.Rectangle, radius int) bool {
	if x < rect.Min.X || x >= rect.Max.X || y < rect.Min.Y || y >= rect.Max.Y {
		return false
	}
	r := float64(radius)
	left := float64(rect.Min.X + radius)
	right := float64(rect.Max.X - radius - 1)
	top := float64(rect.Min.Y + radius)
	bottom := float64(rect.Max.Y - radius - 1)
	px := float64(x)
	py := float64(y)

	switch {
	case px >= left && px <= right:
		return true
	case py >= top && py <= bottom:
		return true
	}

	cx := left
	if px > right {
		cx = right
	}
	cy := top
	if py > bottom {
		cy = bottom
	}

	dx := px - cx
	dy := py - cy
	return dx*dx+dy*dy <= r*r
}

func pointInTriangle(px, py float64, a, b, c image.Point) bool {
	ax, ay := float64(a.X), float64(a.Y)
	bx, by := float64(b.X), float64(b.Y)
	cx, cy := float64(c.X), float64(c.Y)

	v0x, v0y := cx-ax, cy-ay
	v1x, v1y := bx-ax, by-ay
	v2x, v2y := px-ax, py-ay

	dot00 := v0x*v0x + v0y*v0y
	dot01 := v0x*v1x + v0y*v1y
	dot02 := v0x*v2x + v0y*v2y
	dot11 := v1x*v1x + v1y*v1y
	dot12 := v1x*v2x + v1y*v2y

	invDenom := 1 / (dot00*dot11 - dot01*dot01)
	u := (dot11*dot02 - dot01*dot12) * invDenom
	v := (dot00*dot12 - dot01*dot02) * invDenom
	return u >= 0 && v >= 0 && u+v <= 1
}

func lerp(a, b color.NRGBA, t float64) color.NRGBA {
	return color.NRGBA{
		R: uint8(math.Round(float64(a.R) + (float64(b.R)-float64(a.R))*t)),
		G: uint8(math.Round(float64(a.G) + (float64(b.G)-float64(a.G))*t)),
		B: uint8(math.Round(float64(a.B) + (float64(b.B)-float64(a.B))*t)),
		A: uint8(math.Round(float64(a.A) + (float64(b.A)-float64(a.A))*t)),
	}
}

func writePNG(path string, img image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, img)
}

func writeICO(path string, img image.Image) error {
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, img); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, value := range []uint16{0, 1, 1} {
		if err := binary.Write(file, binary.LittleEndian, value); err != nil {
			return err
		}
	}
	if _, err := file.Write([]byte{0, 0, 0, 0}); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, uint16(1)); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, uint16(32)); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, uint32(pngBuf.Len())); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, uint32(22)); err != nil {
		return err
	}
	_, err = file.Write(pngBuf.Bytes())
	return err
}

func min3(a, b, c int) int {
	if a > b {
		a = b
	}
	if a > c {
		a = c
	}
	return a
}

func max3(a, b, c int) int {
	if a < b {
		a = b
	}
	if a < c {
		a = c
	}
	return a
}
