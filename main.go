package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"sync"

	"image-analyzer/analyzers"
)

type AnalysisResult struct {
	FileName     string                 `json:"file_name"`
	FileSize     int64                  `json:"file_size"`
	Width        int                    `json:"width"`
	Height       int                    `json:"height"`
	AspectRatio  string                 `json:"aspect_ratio"`
	Blur         BlurResult             `json:"blur"`
	Brightness   BrightnessResult       `json:"brightness"`
	Noise        NoiseResult            `json:"noise"`
	Exif         analyzers.ExifData     `json:"exif"`
	ColorProfile analyzers.ColorProfile `json:"color_profile"`
	Score        float64                `json:"overall_score"`
}

type BlurResult struct {
	Sharpness float64 `json:"sharpness"`
	Status    string  `json:"status"`
}

type BrightnessResult struct {
	Value  float64 `json:"value"`
	Status string  `json:"status"`
}

type NoiseResult struct {
	Level  float64 `json:"level"`
	Status string  `json:"status"`
}

type BatchResult struct {
	Results []AnalysisResult `json:"results"`
	Total   int              `json:"total"`
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/analyze", handleAnalyze)
	http.HandleFunc("/analyze-batch", handleBatch)
	http.HandleFunc("/export-csv", handleExportCSV)
	fmt.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "File too large. Maximum size is 10MB", http.StatusRequestEntityTooLarge)
		return
	}
	file, header, err := r.FormFile("image")
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
	mimeType := http.DetectContentType(imageData)
	if mimeType != "image/jpeg" && mimeType != "image/png" {
		http.Error(w, "Only JPEG and PNG images are supported", http.StatusBadRequest)
		return
	}
	result := analyzeImage(imageData, header.Filename, header.Size)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handleBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests allowed", http.StatusMethodNotAllowed)
		return
	}
	r.ParseMultipartForm(50 << 20)
	files := r.MultipartForm.File["images"]
	if len(files) == 0 {
		http.Error(w, "No images uploaded", http.StatusBadRequest)
		return
	}
	if len(files) > 20 {
		files = files[:20]
	}
	results := make([]AnalysisResult, len(files))
	errs := make([]error, len(files))
	var wg sync.WaitGroup
	for i, fh := range files {
		wg.Add(1)
		go func(idx int, fileHeader *multipart.FileHeader) {
			defer wg.Done()
			f, err := fileHeader.Open()
			if err != nil {
				errs[idx] = fmt.Errorf("could not open %s: %w", fileHeader.Filename, err)
				return
			}
			defer f.Close()
			data, err := io.ReadAll(f)
			if err != nil {
				errs[idx] = fmt.Errorf("could not read %s: %w", fileHeader.Filename, err)
				return
			}
			results[idx] = analyzeImage(data, fileHeader.Filename, fileHeader.Size)
		}(i, fh)
	}
	wg.Wait()
	var validResults []AnalysisResult
	for i, err := range errs {
		if err != nil {
			fmt.Printf("⚠️  Skipped image: %v\n", err)
		} else {
			validResults = append(validResults, results[i])
		}
	}
	if len(validResults) == 0 {
		http.Error(w, "All images failed to process", http.StatusUnprocessableEntity)
		return
	}
	batch := BatchResult{Results: validResults, Total: len(validResults)}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(batch)
}

func handleExportCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests allowed", http.StatusMethodNotAllowed)
		return
	}
	var results []AnalysisResult
	if err := json.NewDecoder(r.Body).Decode(&results); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="image-analysis.csv"`)
	writer := csv.NewWriter(w)
	defer writer.Flush()
	writer.Write([]string{
		"File Name", "File Size (KB)", "Width", "Height", "Aspect Ratio",
		"Sharpness", "Blur Status", "Brightness", "Brightness Status",
		"Noise Level", "Noise Status", "Camera", "ISO", "Focal Length",
		"Exposure Time", "Overall Score",
	})
	for _, r := range results {
		writer.Write([]string{
			r.FileName,
			fmt.Sprintf("%.1f", float64(r.FileSize)/1024),
			strconv.Itoa(r.Width), strconv.Itoa(r.Height),
			r.AspectRatio,
			fmt.Sprintf("%.2f", r.Blur.Sharpness), r.Blur.Status,
			fmt.Sprintf("%.2f", r.Brightness.Value), r.Brightness.Status,
			fmt.Sprintf("%.2f", r.Noise.Level), r.Noise.Status,
			r.Exif.Camera, r.Exif.ISO, r.Exif.FocalLength, r.Exif.ExposureTime,
			fmt.Sprintf("%.0f", r.Score),
		})
	}
}

func analyzeImage(imageData []byte, fileName string, fileSize int64) AnalysisResult {
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return AnalysisResult{FileName: fileName, FileSize: fileSize}
	}
	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y
	aspectRatio := calcAspectRatio(width, height)

	type blurRes struct {
		sharpness float64
		status    string
	}
	type brightRes struct {
		value  float64
		status string
	}
	type noiseRes struct {
		level  float64
		status string
	}

	var wg sync.WaitGroup
	var blur blurRes
	var bright brightRes
	var noise noiseRes
	var exifData analyzers.ExifData
	var colorProfile analyzers.ColorProfile

	wg.Add(5)
	go func() { defer wg.Done(); blur.sharpness, blur.status = analyzers.DetectBlur(img) }()
	go func() { defer wg.Done(); bright.value, bright.status = analyzers.CheckBrightness(img) }()
	go func() { defer wg.Done(); noise.level, noise.status = analyzers.DetectNoise(img) }()
	go func() { defer wg.Done(); exifData = analyzers.ReadExif(imageData) }()
	go func() { defer wg.Done(); colorProfile = analyzers.AnalyzeColor(img) }()
	wg.Wait()

	score := calculateScore(blur.sharpness, bright.value, noise.level)
	return AnalysisResult{
		FileName: fileName, FileSize: fileSize,
		Width: width, Height: height, AspectRatio: aspectRatio,
		Blur:         BlurResult{Sharpness: blur.sharpness, Status: blur.status},
		Brightness:   BrightnessResult{Value: bright.value, Status: bright.status},
		Noise:        NoiseResult{Level: noise.level, Status: noise.status},
		Exif:         exifData,
		ColorProfile: colorProfile,
		Score:        score,
	}
}

func calculateScore(sharpness, brightness, noise float64) float64 {
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
	if noise > 5.0 {
		score -= 20
	} else if noise > 2.5 {
		score -= 10
	}
	if score < 0 {
		score = 0
	}
	return score
}

func calcAspectRatio(width, height int) string {
	if width == 0 || height == 0 {
		return "unknown"
	}
	g := gcd(width, height)
	return fmt.Sprintf("%d:%d", width/g, height/g)
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
