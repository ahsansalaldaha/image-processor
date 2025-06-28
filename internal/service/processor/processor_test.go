package processor

import (
	"image"
	"image/color"
	"testing"
)

func TestGrayscale(t *testing.T) {
	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Fill with a known color
	testColor := color.RGBA{255, 128, 64, 255}
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, testColor)
		}
	}

	processor := NewImageProcessor()
	grayscaleImg := processor.Grayscale(img)

	// Check that the image is grayscale
	bounds := grayscaleImg.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := grayscaleImg.At(x, y).RGBA()
			// In grayscale, R, G, and B should be equal
			if r != g || g != b {
				t.Errorf("Pixel at (%d, %d) is not grayscale: R=%d, G=%d, B=%d", x, y, r, g, b)
			}
		}
	}
}

func TestDownloadImage(t *testing.T) {
	processor := NewImageProcessor()

	// Test with a valid image URL (you might want to use a test image)
	// This test requires internet connection and might fail in CI
	// In a real test environment, you'd mock the HTTP client

	// For now, let's test the error case with an invalid URL
	_, _, err := processor.DownloadImage(nil, "invalid-url")
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}

func TestImageProcessingPipeline(t *testing.T) {
	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))

	// Fill with a test pattern
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			img.Set(x, y, color.RGBA{uint8(x * 5), uint8(y * 5), 128, 255})
		}
	}

	processor := NewImageProcessor()

	// Test the full pipeline
	grayscaleImg := processor.Grayscale(img)

	// Verify the result
	bounds := grayscaleImg.Bounds()
	if bounds.Dx() != 50 || bounds.Dy() != 50 {
		t.Errorf("Expected image size 50x50, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Check that it's actually grayscale
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 10 { // Sample every 10th pixel
		for x := bounds.Min.X; x < bounds.Max.X; x += 10 {
			r, g, b, _ := grayscaleImg.At(x, y).RGBA()
			if r != g || g != b {
				t.Errorf("Pixel at (%d, %d) is not grayscale: R=%d, G=%d, B=%d", x, y, r, g, b)
			}
		}
	}
}
