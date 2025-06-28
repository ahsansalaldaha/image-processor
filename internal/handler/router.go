package handler

import (
	"encoding/json"
	"net/http"

	"image-processing-system/internal/domain"
	"image-processing-system/pkg/message"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/streadway/amqp"
)

func NewRouter(ch *amqp.Channel) http.Handler {
	r := chi.NewRouter()
	r.Use(httprate.LimitByIP(50, 1)) // 50 req/sec

	r.Post("/submit", func(w http.ResponseWriter, r *http.Request) {
		var job domain.ImageJob
		if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		traceID := r.Header.Get("X-Trace-ID")
		encoded, _ := message.Encode(traceID, "url-ingestor", job)

		err := ch.Publish("", "image.urls", false, false, amqp.Publishing{
			ContentType: "application/json",
			Body:        encoded,
		})
		if err != nil {
			http.Error(w, "publish failed", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})

	return r
}
