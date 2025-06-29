package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"image-processing-system/internal/models"

	amqp "github.com/rabbitmq/amqp091-go"
)

// MockChannel is a mock implementation of ChannelInterface for testing
type MockChannel struct {
	closed bool
}

func (m *MockChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	if m.closed {
		return amqp.ErrClosed
	}
	return nil
}

func (m *MockChannel) IsClosed() bool {
	return m.closed
}

func (m *MockChannel) Close() error {
	m.closed = true
	return nil
}

func TestHealthEndpoint(t *testing.T) {
	// Create a mock channel
	ch := &MockChannel{}

	router := NewRouter(ch)
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if response["status"] != "healthy" {
		t.Errorf("expected status 'healthy', got %v", response["status"])
	}

	if response["service"] != "url-ingestor" {
		t.Errorf("expected service 'url-ingestor', got %v", response["service"])
	}
}

func TestSubmitEndpoint(t *testing.T) {
	// Create a mock channel
	ch := &MockChannel{}

	router := NewRouter(ch)

	// Test valid request
	job := models.ImageJob{
		URLs: []string{"http://example.com/image1.jpg", "http://example.com/image2.jpg"},
	}
	jobBytes, _ := json.Marshal(job)

	req, err := http.NewRequest("POST", "/submit", bytes.NewBuffer(jobBytes))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusAccepted {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusAccepted)
	}
}

func TestSubmitEndpointWithClosedChannel(t *testing.T) {
	// Create a mock channel that is closed
	ch := &MockChannel{closed: true}

	router := NewRouter(ch)

	// Test valid request
	job := models.ImageJob{
		URLs: []string{"http://example.com/image1.jpg", "http://example.com/image2.jpg"},
	}
	jobBytes, _ := json.Marshal(job)

	req, err := http.NewRequest("POST", "/submit", bytes.NewBuffer(jobBytes))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should return 500 when channel is closed
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}

func TestStatusEndpoint(t *testing.T) {
	// Create a mock channel
	ch := &MockChannel{}

	router := NewRouter(ch)
	req, err := http.NewRequest("GET", "/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if response["service"] != "url-ingestor" {
		t.Errorf("expected service 'url-ingestor', got %v", response["service"])
	}

	if response["status"] != "running" {
		t.Errorf("expected status 'running', got %v", response["status"])
	}
}

func TestStatsEndpoint(t *testing.T) {
	// Create a mock channel
	ch := &MockChannel{}

	router := NewRouter(ch)
	req, err := http.NewRequest("GET", "/stats", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if response["service"] != "url-ingestor" {
		t.Errorf("expected service 'url-ingestor', got %v", response["service"])
	}
}
