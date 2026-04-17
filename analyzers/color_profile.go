package analyzers

import (
	"image"
	"math"
)

type ColorProfile struct {
	Dominant string  `json:"dominant"` // "red", "green", "blue", "neutral"
	Colorful bool    `json:"colorful"` // is the image vibrant or dull?
	Vibrance float64 `json:"vibrance"` // 0–100, how colorful overall
	AvgR     float64 `json:"avg_r"`    // average red channel 0–255
	AvgG     float64 `json:"avg_g"`    // average green channel 0–255
	AvgB     float64 `json:"avg_b"`    // average blue channel 0–255
}

func AnalyzeColor(img image.Image) ColorProfile {
	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	var sumR, sumG, sumB float64
	count := float64(width * height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			sumR += float64(r) / 257.0 // convert 0–65535 to 0–255
			sumG += float64(g) / 257.0
			sumB += float64(b) / 257.0
		}
	}

	avgR := sumR / count
	avgG := sumG / count
	avgB := sumB / count

	// Dominant color — whichever channel is highest
	dominant := "neutral"
	maxDiff := 15.0 // minimum difference to call it dominant
	if avgR-avgG > maxDiff && avgR-avgB > maxDiff {
		dominant = "red"
	} else if avgG-avgR > maxDiff && avgG-avgB > maxDiff {
		dominant = "green"
	} else if avgB-avgR > maxDiff && avgB-avgG > maxDiff {
		dominant = "blue"
	}

	// Vibrance — how spread apart are the RGB channels?
	// High spread = colorful, low spread = gray/neutral
	mean := (avgR + avgG + avgB) / 3
	variance := (math.Pow(avgR-mean, 2) + math.Pow(avgG-mean, 2) + math.Pow(avgB-mean, 2)) / 3
	vibrance := math.Min(100, math.Sqrt(variance)*3)

	return ColorProfile{
		Dominant: dominant,
		Colorful: vibrance > 15,
		Vibrance: math.Round(vibrance*100) / 100,
		AvgR:     math.Round(avgR*100) / 100,
		AvgG:     math.Round(avgG*100) / 100,
		AvgB:     math.Round(avgB*100) / 100,
	}
}
