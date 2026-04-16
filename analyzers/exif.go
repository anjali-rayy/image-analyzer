package analyzers

import (
	"bytes"
	"fmt"

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
		result.Camera = cam.String()
	}
	if iso, err := x.Get(exif.ISOSpeedRatings); err == nil {
		result.ISO = fmt.Sprintf("%v", iso)
	}
	if fl, err := x.Get(exif.FocalLength); err == nil {
		result.FocalLength = fl.String()
	}
	if et, err := x.Get(exif.ExposureTime); err == nil {
		result.ExposureTime = et.String()
	}

	return result
}
