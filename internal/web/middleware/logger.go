package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	colorReset = "\033[0m"
	colorRed   = 31
	colorGreen = 32
	colorGray  = 90
)

func httpColorize(colorCode int, v string) string {
	return fmt.Sprintf("\033[%sm%s%s", strconv.Itoa(colorCode), v, colorReset)
}

func inboundLogHandler(next http.Handler, logger *slog.Logger, env string) http.Handler {
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
			elapsed := time.Since(startTime)
			status := lrw.responseData.status

			if env == "prod" { //nolint:nestif
				// Verbose structured logging for production (Vercel)
				scheme := "http"
				if r.TLS != nil {
					scheme = "https"
				}
				requestURL := fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)
				size := lrw.responseData.size

				requestFields := slog.Group("request",
					slog.String("url", requestURL),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("remoteIP", r.RemoteAddr),
					slog.String("proto", r.Proto),
					slog.Any("headers", headerLogField(r.Header)),
				)

				responseFields := slog.Group("response",
					slog.Int64("elapsedMs", elapsed.Milliseconds()),
					slog.Int("status", status),
					slog.Int("size", size),
				)

				if status >= http.StatusBadRequest {
					logger.Error("request", requestFields, responseFields)
				} else {
					logger.Info("request", requestFields, responseFields)
				}
			} else {
				// Concise logging for development
				statusColor := colorGray
				if status >= http.StatusBadRequest {
					statusColor = colorRed
				}
				msg := fmt.Sprintf("%s %s %s %s",
					httpColorize(colorGreen, r.Method),
					r.URL.Path,
					httpColorize(statusColor, fmt.Sprintf("%d", status)),
					elapsed.Round(time.Millisecond))
				if status >= http.StatusBadRequest {
					logger.Error(msg)
				} else {
					logger.Info(msg)
				}
			}
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
