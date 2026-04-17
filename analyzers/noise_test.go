package analyzers

import (
	"image"
	"image/color"
	"testing"
)

// makeNoisyImage creates an image with random-looking pixel variation
func makeNoisyImage(width, height int) image.Image {
	img := image.NewGray(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Alternate between very dark and very bright — simulates noise
			if (x*y)%3 == 0 {
				img.SetGray(x, y, color.Gray{Y: 10})
			} else if (x*y)%3 == 1 {
				img.SetGray(x, y, color.Gray{Y: 245})
			} else {
				img.SetGray(x, y, color.Gray{Y: 128})
			}
		}
	}
	return img
}

func TestDetectNoise_CleanImage(t *testing.T) {
	// Solid image has no variation between neighbors — should be clean
	img := makeGrayImage(100, 100, 128)
	level, status := DetectNoise(img)

	if status != "clean" {
		t.Errorf("expected 'clean' for solid image, got '%s'", status)
	}
	if level > 2.5 {
		t.Errorf("expected noise level < 2.5 for solid image, got %f", level)
	}
}

func TestDetectNoise_NoisyImage(t *testing.T) {
	// Randomly varying pixels should be noisy
	img := makeNoisyImage(100, 100)
	level, status := DetectNoise(img)

	if status == "clean" {
		t.Errorf("expected 'noisy' or 'moderate' for noisy image, got '%s'", status)
	}
	if level < 2.5 {
		t.Errorf("expected noise level > 2.5 for noisy image, got %f", level)
	}
}

func TestDetectNoise_TooSmall(t *testing.T) {
	img := makeGrayImage(2, 2, 128)
	_, status := DetectNoise(img)

	if status != "unknown" {
		t.Errorf("expected 'unknown' for tiny image, got '%s'", status)
	}
}
