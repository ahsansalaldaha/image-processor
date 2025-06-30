package handler

import (
	"bytes"
	"net/http"
	"testing"
)

const requestBody = `{
  "urls": [
    "https://demonslayer-anime.com/portal/assets/img/img_kv_2.jpg",
    "https://demonslayer-hinokami.sega.com/img/purchase/digital-standard.jpg"
  ]
}`

func BenchmarkSubmitAPI(b *testing.B) {
	client := &http.Client{}
	url := "http://localhost:8080/submit"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, err := http.NewRequest("POST", url, bytes.NewBufferString(requestBody))
			if err != nil {
				b.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err != nil {
				b.Fatal(err)
			}
			resp.Body.Close()
		}
	})
}
