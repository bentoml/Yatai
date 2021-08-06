package utils

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

type LoggingResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (l *LoggingResponseWriter) WriteHeader(code int) {
	l.StatusCode = code
	l.ResponseWriter.WriteHeader(code)
}

func APIOutputJson(ctx context.Context, w http.ResponseWriter, statusCode int, data interface{}) {
	header := w.Header()
	header.Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		logrus.Errorf("failed to outputJson: %v", err)
	}
}

func APIOutputOK(ctx context.Context, w http.ResponseWriter, d interface{}) {
	if o, ok := d.(string); ok {
		d = struct {
			Status  bool
			Message string
		}{true, o}
	}
	APIOutputJson(ctx, w, http.StatusOK, d)
}

type ErrResponse struct {
	Status  bool
	Code    int
	Message string
}

func APIOutputErr(ctx context.Context, w http.ResponseWriter, statusCode int, msg string) {
	APIOutputJson(ctx, w, statusCode, ErrResponse{
		Status:  false,
		Code:    statusCode,
		Message: msg,
	})
}
