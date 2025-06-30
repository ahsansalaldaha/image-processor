package models

type ImageJob struct {
	URLs            []string `json:"urls"`
	ProcessingTypes []string `json:"processing_types"`
}
