package analyzers

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
)

type ExifData struct {
	Camera       string `json:"camera"`
	ISO          string `json:"iso"`
	FocalLength  string `json:"focal_length"`
	ExposureTime string `json:"exposure_time"`
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
		return result
	}

	if cam, err := x.Get(exif.Model); err == nil {
		// cam.String() wraps value in quotes e.g. "Inspiron 14 5440" — strip them
		result.Camera = strings.Trim(cam.String(), "\"")
	}
	if iso, err := x.Get(exif.ISOSpeedRatings); err == nil {
		// iso.String() returns "[800]" style — use StringVal for clean output
		result.ISO = strings.TrimSpace(strings.Trim(iso.String(), "[]"))
	}
	if fl, err := x.Get(exif.FocalLength); err == nil {
		num, err1 := fl.Rat(0)
		if err1 == nil {
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
