package web

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a-h/templ"
)

type committedResponseRecorder struct {
	*httptest.ResponseRecorder
}

func (r committedResponseRecorder) ResponseCommitted() bool {
	return true
}

func TestAppHandlerDoesNotRenderErrorAfterRequestCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	recorder := httptest.NewRecorder()
	handler := NewAppHandler(slog.New(slog.DiscardHandler))

	handler.Handle(func(reqResp *ReqRespPair) error {
		reqResp.Response.WriteHeader(http.StatusOK)
		return context.Canceled
	})(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("response status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("unexpected error response: %s", recorder.Body.String())
	}
}

func TestAppHandlerRendersDeadlineErrorWhileRequestIsActive(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()
	handler := NewAppHandler(slog.New(slog.DiscardHandler))

	handler.Handle(func(_ *ReqRespPair) error {
		return context.DeadlineExceeded
	})(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("response status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if recorder.Body.Len() == 0 {
		t.Fatal("expected an error response body")
	}
}

func TestAppHandlerDoesNotRenderSecondErrorAfterResponseStarted(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := committedResponseRecorder{ResponseRecorder: httptest.NewRecorder()}
	handler := NewAppHandler(slog.New(slog.DiscardHandler))

	handler.Handle(func(reqResp *ReqRespPair) error {
		reqResp.Response.WriteHeader(http.StatusOK)
		return errors.New("write response: i/o timeout")
	})(&recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("response status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("unexpected second response: %s", recorder.Body.String())
	}
}

func TestRenderWritesCompleteHTMLResponse(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()
	reqResp := &ReqRespPair{Request: req, Response: recorder}
	component := templ.ComponentFunc(func(_ context.Context, writer io.Writer) error {
		_, err := io.WriteString(writer, "<p>ready</p>")
		return err
	})

	err := reqResp.Render(component)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if recorder.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Fatalf("content type = %q", recorder.Header().Get("Content-Type"))
	}
	if recorder.Header().Get("Content-Length") != "12" {
		t.Fatalf("content length = %q, want 12", recorder.Header().Get("Content-Length"))
	}
	if recorder.Body.String() != "<p>ready</p>" {
		t.Fatalf("response body = %q", recorder.Body.String())
	}
}
