package analyzers

import (
	"testing"
)

func TestCheckBrightness_TooDark(t *testing.T) {
	// Pure black image should be "too dark"
	img := makeGrayImage(100, 100, 0)
	brightness, status := CheckBrightness(img)

	if status != "too dark" {
		t.Errorf("expected 'too dark' for black image, got '%s'", status)
	}
	if brightness > 20 {
		t.Errorf("expected brightness < 20 for black image, got %f", brightness)
	}
}

func TestCheckBrightness_TooBright(t *testing.T) {
	// Pure white image should be "too bright"
	img := makeGrayImage(100, 100, 255)
	brightness, status := CheckBrightness(img)

	if status != "too bright" {
		t.Errorf("expected 'too bright' for white image, got '%s'", status)
	}
	if brightness < 80 {
		t.Errorf("expected brightness > 80 for white image, got %f", brightness)
	}
}

func TestCheckBrightness_Good(t *testing.T) {
	// Mid-gray image should be "good"
	img := makeGrayImage(100, 100, 128)
	_, status := CheckBrightness(img)

	if status != "good" {
		t.Errorf("expected 'good' for mid-gray image, got '%s'", status)
	}
}
