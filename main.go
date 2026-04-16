package main

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"

	"image-analyzer/analyzers"
)

type AnalysisResult struct {
	Blur       BlurResult         `json:"blur"`
	Brightness BrightnessResult   `json:"brightness"`
	Exif       analyzers.ExifData `json:"exif"`
	Score      float64            `json:"overall_score"`
}

type BlurResult struct {
	Sharpness float64 `json:"sharpness"`
	Status    string  `json:"status"`
}

type BrightnessResult struct {
	Value  float64 `json:"value"`
	Status string  `json:"status"`
}

func main() {
	http.HandleFunc("/analyze", handleAnalyze)
	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(10 << 20)
	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Could not read image: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	imageData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Could not read image data", http.StatusInternalServerError)
		return
	}

	img, _, err := image.Decode(bytesReader(imageData))
	if err != nil {
		http.Error(w, "Could not decode image", http.StatusBadRequest)
		return
	}

	sharpness, blurStatus := analyzers.DetectBlur(img)
	brightness, brightStatus := analyzers.CheckBrightness(img)
	exifData := analyzers.ReadExif(imageData)

	score := calculateScore(sharpness, brightness)

	result := AnalysisResult{
		Blur:       BlurResult{Sharpness: sharpness, Status: blurStatus},
		Brightness: BrightnessResult{Value: brightness, Status: brightStatus},
		Exif:       exifData,
		Score:      score,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func calculateScore(sharpness, brightness float64) float64 {
	score := 100.0
	if sharpness < 3.0 {
		score -= 40
	} else if sharpness < 6.0 {
		score -= 20
	}
	if brightness < 20 || brightness > 80 {
		score -= 30
	} else if brightness < 30 || brightness > 70 {
		score -= 10
	}
	if score < 0 {
		score = 0
	}
	return score
}

func bytesReader(data []byte) *bytesReaderImpl {
	return &bytesReaderImpl{data: data, pos: 0}
}

type bytesReaderImpl struct {
	data []byte
	pos  int
}

func (b *bytesReaderImpl) Read(p []byte) (n int, err error) {
	if b.pos >= len(b.data) {
		return 0, io.EOF
	}
	n = copy(p, b.data[b.pos:])
	b.pos += n
	return n, nil
}

func (b *bytesReaderImpl) Seek(offset int64, whence int) (int64, error) {
	var newPos int64
	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = int64(b.pos) + offset
	case io.SeekEnd:
		newPos = int64(len(b.data)) + offset
	}
	b.pos = int(newPos)
	return newPos, nil
}
