package processor

import (
	"context"
	"image"
	"net/http"

	"github.com/disintegration/imaging"
)

func DownloadImage(ctx context.Context, url string) (image.Image, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	return image.Decode(resp.Body)
}

func Grayscale(img image.Image) image.Image {
	return imaging.Grayscale(img)
}
