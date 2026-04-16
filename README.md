# Image Quality Analyzer API

A REST API built in Go that analyzes image quality by detecting blur, brightness, and EXIF metadata.

## What it does

Send any image → get back a JSON quality report with an overall score.

## Tech Stack

- **Language:** Go
- **Libraries:** `golang.org/x/image`, `goexif`
- **Architecture:** Single endpoint, 3 concurrent analyzers

## API Usage

**Endpoint:** `POST /analyze`

**Request:** `multipart/form-data` with an `image` field (JPEG or PNG)

**Response:**
```json
{
  "blur":       { "sharpness": 24.78, "status": "sharp" },
  "brightness": { "value": 76.41,    "status": "good"  },
  "exif":       { "camera": "unknown", "iso": "unknown" },
  "overall_score": 90
}
```

## How to Run

```bash
git clone https://github.com/anjali-rayy/image-analyzer
cd image-analyzer
go run main.go
```

Then send a POST request to `http://localhost:8080/analyze` with an image file.

## Analyzers

- **Blur** — Laplacian variance method to detect sharpness
- **Brightness** — Average pixel luminance (flags too dark / too bright)
- **EXIF** — Extracts camera model, ISO, focal length, exposure time