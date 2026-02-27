package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func inboundLogHandler(next http.Handler, logger *slog.Logger) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/assets") {
			next.ServeHTTP(w, r)
			return
		}
		startTime := time.Now()
		lrw := loggingResponseWriter{
			wrapped:      w,
			responseData: &responseData{},
		}

		defer func() {
			endTime := time.Since(startTime)
			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			requestURL := fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)
			status := lrw.responseData.status
			size := lrw.responseData.size
			msg := fmt.Sprintf("Incoming request %s: %d %s", r.URL.Path, status, statusLabel(status))

			requestFields := slog.Group("request",
				slog.String("url", requestURL),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remoteIP", r.RemoteAddr),
				slog.String("proto", r.Proto),
				slog.Any("headers", headerLogField(r.Header)),
			)

			responseFields := slog.Group("response",
				slog.Int64("elapsedMs", endTime.Milliseconds()),
				slog.Int("status", status),
				slog.Int("size", size),
			)

			logger.Info(msg, requestFields, responseFields)
		}()

		next.ServeHTTP(&lrw, r)
	}

	return http.HandlerFunc(fn)
}

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	wrapped      http.ResponseWriter
	responseData *responseData
}

func (s *loggingResponseWriter) Header() http.Header {
	return s.wrapped.Header()
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.wrapped.Write(b)
	r.responseData.size += size
	if err != nil {
		return size, fmt.Errorf("write response: %w", err)
	}
	return size, nil
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.wrapped.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func statusLabel(status int) string {
	switch {
	case status >= 100 && status < 300:
		return "OK"
	case status >= 300 && status < 400:
		return "Redirect"
	case status >= 400 && status < 500:
		return "Client Error"
	case status >= 500: //nolint:mnd
		return "Server Error"
	default:
		return "Unknown"
	}
}

func headerLogField(header http.Header) []slog.Attr {
	headerField := []slog.Attr{}
	for key, values := range header {
		key = strings.ToLower(key)
		switch {
		case len(values) == 0:
			continue
		case len(values) == 1:
			headerField = append(headerField, slog.Attr{Key: key, Value: slog.StringValue(values[0])})
		default:
			headerField = append(headerField, slog.Attr{Key: key, Value: slog.StringValue(fmt.Sprintf("[%s]", strings.Join(values, "], [")))})
		}
		if key == "authorization" || key == "cookie" || key == "set-cookie" {
			headerField[len(headerField)-1] = slog.Attr{
				Key:   key,
				Value: slog.StringValue("***"),
			}
		}
	}
	return headerField
}
