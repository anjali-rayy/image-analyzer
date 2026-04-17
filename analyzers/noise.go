package analyzers

import (
	"image"
	"math"
)

// DetectNoise estimates noise by comparing each pixel to its neighbors.
// High-frequency random variation = noise.
func DetectNoise(img image.Image) (float64, string) {
	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	if width < 3 || height < 3 {
		return 0, "unknown"
	}

	var totalDiff float64
	count := 0

	// For each pixel (excluding borders), compare to its 4 neighbors.
	// If a pixel is very different from neighbors in a random way, that's noise.
	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			gray := toGray(img, x, y)

			// Get 4 neighbors
			top := toGray(img, x, y-1)
			bottom := toGray(img, x, y+1)
			left := toGray(img, x-1, y)
			right := toGray(img, x+1, y)

			// Laplacian: how different is this pixel from its neighbors?
			laplacian := math.Abs(4*gray - top - bottom - left - right)
			totalDiff += laplacian
			count++
		}
	}

	noiseLevel := (totalDiff / float64(count)) * 100

	// Round to 2 decimal places
	noiseLevel = math.Round(noiseLevel*100) / 100

	status := "clean"
	if noiseLevel > 5.0 {
		status = "noisy"
	} else if noiseLevel > 2.5 {
		status = "moderate"
	}

	return noiseLevel, status
}

func toGray(img image.Image, x, y int) float64 {
	r, g, b, _ := img.At(x, y).RGBA()
	return (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 65535.0
}
