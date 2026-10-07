package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggingResponseWriterIgnoresSecondStatus(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	data := &responseData{}
	writer := &loggingResponseWriter{wrapped: recorder, responseData: data}

	writer.WriteHeader(http.StatusOK)
	writer.WriteHeader(http.StatusInternalServerError)

	if recorder.Code != http.StatusOK {
		t.Fatalf("response status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if data.status != http.StatusOK {
		t.Fatalf("logged status = %d, want %d", data.status, http.StatusOK)
	}
}
