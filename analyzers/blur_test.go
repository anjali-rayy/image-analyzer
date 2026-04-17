package analyzers

import (
	"image"
	"image/color"
	"testing"
)

// makeGrayImage creates a solid gray image for testing
func makeGrayImage(width, height int, grayValue uint8) image.Image {
	img := image.NewGray(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetGray(x, y, color.Gray{Y: grayValue})
		}
	}
	return img
}

// makeCheckerImage creates a black/white checkerboard — very sharp edges
func makeCheckerImage(width, height int) image.Image {
	img := image.NewGray(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if (x+y)%2 == 0 {
				img.SetGray(x, y, color.Gray{Y: 255})
			} else {
				img.SetGray(x, y, color.Gray{Y: 0})
			}
		}
	}
	return img
}

func TestDetectBlur_SolidImage(t *testing.T) {
	// A solid color image has no edges — should be blurry
	img := makeGrayImage(100, 100, 128)
	sharpness, status := DetectBlur(img)

	if status != "blurry" {
		t.Errorf("expected 'blurry' for solid image, got '%s'", status)
	}
	if sharpness != 0 {
		t.Errorf("expected sharpness 0 for solid image, got %f", sharpness)
	}
}

func TestDetectBlur_SharpImage(t *testing.T) {
	// Checkerboard has maximum edges — should be sharp
	img := makeCheckerImage(100, 100)
	sharpness, status := DetectBlur(img)

	if status != "sharp" {
		t.Errorf("expected 'sharp' for checkerboard image, got '%s'", status)
	}
	if sharpness <= 3.0 {
		t.Errorf("expected sharpness > 3.0 for checkerboard, got %f", sharpness)
	}
}

func TestDetectBlur_TooSmall(t *testing.T) {
	// Image smaller than 3x3 should return unknown
	img := makeGrayImage(2, 2, 128)
	_, status := DetectBlur(img)

	if status != "unknown" {
		t.Errorf("expected 'unknown' for tiny image, got '%s'", status)
	}
}
