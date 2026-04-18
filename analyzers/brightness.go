package analyzers

import (
	"image"
	"math"
)

func CheckBrightness(img image.Image) (float64, string) {
	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	var totalBrightness float64
	count := float64(width * height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			brightness := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 65535.0 * 100
			totalBrightness += brightness
		}
	}

	avg := totalBrightness / count

	status := "good"
	if avg < 20 {
		status = "too dark"
	} else if avg > 80 {
		status = "too bright"
	}

	return math.Round(avg*100) / 100, status
}
