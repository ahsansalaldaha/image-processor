package main

import (
	"image-processing-system/internal/metadata"
	"image-processing-system/pkg/rabbitmq"
	"image-processing-system/pkg/tracing"
	"log"
)

func main() {
	tracer := tracing.Init("image-metadata")
	defer tracer.Shutdown()

	db := metadata.InitDB()

	defer db.DB().Close()

	conn, ch := rabbitmq.Connect()
	defer conn.Close()
	defer ch.Close()

	log.Println("image-metadata service consuming processed image queue")
	metadata.ConsumeAndStore(ch, db)
}
