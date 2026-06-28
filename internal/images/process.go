// Package images normalizes uploaded photos: it strips metadata (EXIF, including
// GPS) by re-encoding to JPEG, and renders a thumbnail.
package images

import (
	"bytes"
	"fmt"
	_ "golang.org/x/image/webp" // register WEBP decoder
	"image"
	"image/jpeg"
	_ "image/png" // register PNG decoder

	"github.com/disintegration/imaging"
)

// Processed is the normalized output for one uploaded image.
type Processed struct {
	Full  []byte // full-size JPEG, metadata stripped
	Thumb []byte // thumbnail JPEG
}

// Process decodes raw image bytes (JPEG/PNG/WEBP), re-encodes the full image as
// JPEG (which drops EXIF and any other metadata), and renders a thumbnail that
// fits within thumbMax × thumbMax.
func Process(raw []byte, thumbMax int) (*Processed, error) {
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	var full bytes.Buffer
	if err := jpeg.Encode(&full, img, &jpeg.Options{Quality: 90}); err != nil {
		return nil, fmt.Errorf("encode full: %w", err)
	}

	thumbImg := imaging.Fit(img, thumbMax, thumbMax, imaging.Lanczos)
	var thumb bytes.Buffer
	if err := jpeg.Encode(&thumb, thumbImg, &jpeg.Options{Quality: 80}); err != nil {
		return nil, fmt.Errorf("encode thumb: %w", err)
	}

	return &Processed{Full: full.Bytes(), Thumb: thumb.Bytes()}, nil
}
