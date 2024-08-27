package models

import "github.com/sebasegovia01/base-template-go-gin/enums"

type ErrorResponse struct {
	Result ResultError `json:"Result"`
}

type ResultError struct {
	Status         enums.ResultStatus `json:"status"`
	Description    string             `json:"description,omitempty"`
	CanonicalError *CanonicalError    `json:"CanonicalError,omitempty"`
	SourceError    *SourceError       `json:"SourceError,omitempty"`
}

type CanonicalError struct {
	Code        string                   `json:"code"`
	Type        enums.CanonicalErrorType `json:"type"`
	Description string                   `json:"description"`
}

type SourceError struct {
	Code               string             `json:"code"`
	Description        string             `json:"description"`
	ErrorSourceDetails ErrorSourceDetails `json:"ErrorSourceDetails"`
}

type ErrorSourceDetails struct {
	Source         string   `json:"source"`
	MissingHeaders []string `json:"missingHeaders,omitempty"`
}
