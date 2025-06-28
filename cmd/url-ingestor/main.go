package main

import (
	"image-processing-system/internal/handler"
	"image-processing-system/pkg/auth"
	"image-processing-system/pkg/rabbitmq"
	"image-processing-system/pkg/tracing"
	"log"
	"net/http"
)

func main() {
	tracer := tracing.Init("url-ingestor")
	defer tracer.Shutdown()

	conn, ch := rabbitmq.Connect()
	defer conn.Close()
	defer ch.Close()

	router := handler.NewRouter(ch)
	tlsCfg, err := auth.LoadMutualTLSConfig("./docker/certs/server.crt", "./docker/certs/server.key", "./docker/certs/ca.crt")
	if err != nil {
		log.Fatalf("TLS config error: %v", err)
	}

	srv := &http.Server{
		Addr:      ":8080",
		Handler:   router,
		TLSConfig: tlsCfg,
	}

	log.Println("url-ingestor listening on :8080")
	log.Fatal(srv.ListenAndServeTLS("", ""))
}
