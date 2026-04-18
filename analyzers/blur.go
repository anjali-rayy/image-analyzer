package analyzers

import (
	"image"
	"math"
)

// DetectBlur measures sharpness using Laplacian variance.
// High variance = sharp edges present = sharp image.
func DetectBlur(img image.Image) (float64, string) {
	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	if width < 3 || height < 3 {
		return 0, "unknown"
	}

	var sum, sumSq float64
	count := 0

	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			gray := toGray(img, x, y)
			top := toGray(img, x, y-1)
			bottom := toGray(img, x, y+1)
			left := toGray(img, x-1, y)
			right := toGray(img, x+1, y)

			lap := 4*gray - top - bottom - left - right
			sum += lap
			sumSq += lap * lap
			count++
		}
	}

	c := float64(count)
	mean := sum / c
	variance := (sumSq / c) - (mean * mean)
	sharpness := math.Sqrt(math.Abs(variance))

	status := "sharp"
	if sharpness < 0.03 {
		status = "blurry"
	}

	return math.Round(sharpness*100) / 100, status
}
