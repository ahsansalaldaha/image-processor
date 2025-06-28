package storage

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var client *minio.Client

func init() {
	client, _ = minio.New("minio:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("minioadmin", "minioadmin", ""),
		Secure: false,
	})
	client.MakeBucket(context.Background(), "images", minio.MakeBucketOptions{})
}

func UploadToMinio(ctx context.Context, img image.Image) error {
	buf := new(bytes.Buffer)
	_ = jpeg.Encode(buf, img, nil)
	_, err := client.PutObject(ctx, "images", time.Now().Format("20060102150405")+".jpg", bytes.NewReader(buf.Bytes()), int64(buf.Len()), minio.PutObjectOptions{ContentType: "image/jpeg"})
	return err
}
