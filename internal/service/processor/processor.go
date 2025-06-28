package processor

import (
	"context"
	"fmt"
	"image"
	"net/http"
	"time"

	"github.com/disintegration/imaging"
)

// ImageProcessor handles image processing operations
type ImageProcessor struct {
	client *http.Client
}

// NewImageProcessor creates a new image processor instance
func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// DownloadImage downloads an image from a URL
func (p *ImageProcessor) DownloadImage(ctx context.Context, url string) (image.Image, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	img, format, err := image.Decode(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image: %w", err)
	}

	return img, format, nil
}

// Grayscale converts an image to grayscale
func (p *ImageProcessor) Grayscale(img image.Image) image.Image {
	return imaging.Grayscale(img)
}

// Resize resizes an image to the specified dimensions
func (p *ImageProcessor) Resize(img image.Image, width, height int) image.Image {
	return imaging.Resize(img, width, height, imaging.Lanczos)
}

// Blur applies a blur effect to an image
func (p *ImageProcessor) Blur(img image.Image, sigma float64) image.Image {
	return imaging.Blur(img, sigma)
}

// Sharpen applies a sharpen effect to an image
func (p *ImageProcessor) Sharpen(img image.Image, sigma float64) image.Image {
	return imaging.Sharpen(img, sigma)
}
