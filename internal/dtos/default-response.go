package dtos

import (
	"time"

	"github.com/oklog/ulid/v2"
)

type defaultResponse struct {
	RequestId string `json:"requestId"`
	Message string `json:"message"`
	Data any `json:"data"`
	Timestamp time.Time `json:"timestamp"`
}

func NewDefaultResponse(message string, data any) defaultResponse {
	return defaultResponse{
		RequestId: ulid.Make().String(),
		Message: message,
		Data: data,
		Timestamp: time.Now(),
	}
}