package analyzers

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
)

type ExifData struct {
	Camera       string `json:"camera"`
	ISO          string `json:"iso"`
	FocalLength  string `json:"focal_length"`
	ExposureTime string `json:"exposure_time"`
}

func init() {
	// Register makernote parsers — required for many phones (Samsung, Apple, etc.)
	exif.RegisterParsers(mknote.All...)
}

func ReadExif(imageData []byte) ExifData {
	result := ExifData{
		Camera:       "unknown",
		ISO:          "unknown",
		FocalLength:  "unknown",
		ExposureTime: "unknown",
	}

	x, err := exif.Decode(bytes.NewReader(imageData))
	if err != nil {
		// Try walking raw bytes to find EXIF marker manually
		result.Camera = extractCameraFallback(imageData)
		return result
	}

	// Try Make + Model combined (e.g. "Samsung SM-G991B")
	var make_, model_ string
	if mk, err := x.Get(exif.Make); err == nil {
		make_ = strings.Trim(mk.String(), "\"")
	}
	if md, err := x.Get(exif.Model); err == nil {
		model_ = strings.Trim(md.String(), "\"")
	}

	if make_ != "" && model_ != "" {
		// Avoid duplication like "Samsung Samsung Galaxy S21"
		if strings.HasPrefix(model_, make_) {
			result.Camera = model_
		} else {
			result.Camera = make_ + " " + model_
		}
	} else if model_ != "" {
		result.Camera = model_
	} else if make_ != "" {
		result.Camera = make_
	}

	if iso, err := x.Get(exif.ISOSpeedRatings); err == nil {
		result.ISO = strings.TrimSpace(strings.Trim(iso.String(), "[]"))
	}

	if fl, err := x.Get(exif.FocalLength); err == nil {
		if num, err := fl.Rat(0); err == nil {
			f, _ := num.Float64()
			result.FocalLength = fmt.Sprintf("%.0fmm", f)
		} else {
			result.FocalLength = fl.String()
		}
	}

	if et, err := x.Get(exif.ExposureTime); err == nil {
		result.ExposureTime = et.String() + "s"
	}

	return result
}

// extractCameraFallback scans raw bytes for ASCII camera strings
// when EXIF decode fails (common with some Android/iOS image variants)
func extractCameraFallback(data []byte) string {
	// Known camera brand prefixes to scan for
	brands := []string{
		"Apple", "Samsung", "Google", "Xiaomi", "OnePlus", "Oppo", "Vivo",
		"Huawei", "Sony", "Canon", "Nikon", "Motorola", "Realme", "Nokia",
	}

	dataStr := string(data)
	for _, brand := range brands {
		idx := strings.Index(dataStr, brand)
		if idx != -1 {
			// Extract up to 40 chars from brand start, clean non-printable chars
			end := idx + 40
			if end > len(dataStr) {
				end = len(dataStr)
			}
			raw := dataStr[idx:end]
			var clean strings.Builder
			for _, c := range raw {
				if c >= 32 && c < 127 {
					clean.WriteRune(c)
				} else {
					break
				}
			}
			result := strings.TrimSpace(clean.String())
			if len(result) > 3 {
				return result
			}
		}
	}
	return "unknown"
}
