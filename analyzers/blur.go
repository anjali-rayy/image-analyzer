package analyzers

import (
	"image"
	"math"
)

func DetectBlur(img image.Image) (float64, string) {
	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	var variance float64
	var sum, sumSq float64
	count := float64(width * height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			gray := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 65535.0
			sum += gray
			sumSq += gray * gray
		}
	}

	mean := sum / count
	variance = (sumSq / count) - (mean * mean)
	sharpness := math.Sqrt(variance) * 100

	status := "sharp"
	if sharpness < 3.0 {
		status = "blurry"
	}

	return math.Round(sharpness*100) / 100, status
}
