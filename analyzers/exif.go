package analyzers

import (
	"bytes"
	"fmt"
	"strings"

	exifcommon "github.com/dsoprea/go-exif/v3/common"
	jpegstructure "github.com/dsoprea/go-jpeg-image-structure/v2"
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
	exif.RegisterParsers(mknote.All...)
}

func ReadExif(imageData []byte) ExifData {
	result := ExifData{
		Camera:       "unknown",
		ISO:          "unknown",
		FocalLength:  "unknown",
		ExposureTime: "unknown",
	}

	// Strategy 1: try rwcarlsen/goexif (fast, works for most laptop/DSLR images)
	if tryGoExif(imageData, &result) {
		return result
	}

	// Strategy 2: try dsoprea/go-jpeg-image-structure (better for phone JPEGs)
	tryDsopreaExif(imageData, &result)

	return result
}

func tryGoExif(imageData []byte, result *ExifData) bool {
	x, err := exif.Decode(bytes.NewReader(imageData))
	if err != nil {
		return false
	}

	found := false

	var make_, model_ string
	if mk, err := x.Get(exif.Make); err == nil {
		make_ = strings.Trim(mk.String(), "\"")
		found = true
	}
	if md, err := x.Get(exif.Model); err == nil {
		model_ = strings.Trim(md.String(), "\"")
		found = true
	}
	if make_ != "" && model_ != "" {
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
		}
	}
	if et, err := x.Get(exif.ExposureTime); err == nil {
		result.ExposureTime = et.String() + "s"
	}

	return found
}

func tryDsopreaExif(imageData []byte, result *ExifData) {
	defer func() { recover() }() // dsoprea can panic on malformed data

	jmp := jpegstructure.NewJpegMediaParser()
	intfc, err := jmp.ParseBytes(imageData)
	if err != nil {
		return
	}

	sl := intfc.(*jpegstructure.SegmentList)
	_, _, exifTags, err := sl.DumpExif()
	if err != nil {
		return
	}

	var make_, model_ string
	for _, tag := range exifTags {
		switch tag.TagName {
		case "Make":
			make_ = fmt.Sprintf("%v", tag.Value)
		case "Model":
			model_ = fmt.Sprintf("%v", tag.Value)
		case "ISOSpeedRatings":
			if result.ISO == "unknown" {
				result.ISO = fmt.Sprintf("%v", tag.Value)
			}
		case "FocalLength":
			if result.FocalLength == "unknown" {
				val := fmt.Sprintf("%v", tag.Value)
				result.FocalLength = val + "mm"
			}
		case "ExposureTime":
			if result.ExposureTime == "unknown" {
				result.ExposureTime = fmt.Sprintf("%v", tag.Value) + "s"
			}
		}
	}

	if make_ != "" || model_ != "" {
		if strings.HasPrefix(model_, make_) {
			result.Camera = model_
		} else if make_ != "" && model_ != "" {
			result.Camera = make_ + " " + model_
		} else {
			result.Camera = make_ + model_
		}
	}
}

// helper kept for reference
func extractCameraFallback(data []byte) string {
	brands := []string{
		"Apple", "Samsung", "Google", "Xiaomi", "OnePlus", "Oppo", "Vivo",
		"Huawei", "Sony", "Canon", "Nikon", "Motorola", "Realme", "Nokia",
	}
	dataStr := string(data)
	for _, brand := range brands {
		idx := strings.Index(dataStr, brand)
		if idx != -1 {
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
			res := strings.TrimSpace(clean.String())
			if len(res) > 3 {
				return res
			}
		}
	}
	return "unknown"
}

// needed by dsoprea
var _ = exifcommon.TypeAscii
