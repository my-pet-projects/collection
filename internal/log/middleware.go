package log

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

func NewLoggingMiddleware(logger *slog.Logger) echo.MiddlewareFunc {
	logger = logger.WithGroup("http")
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			start := time.Now()

			reqAttr := []slog.Attr{
				slog.String("method", req.Method),
				slog.String("url", req.URL.String()),
				slog.String("query", req.URL.RawQuery),
				// slog.String("protocol", r.Proto),
				slog.String("remoteIp", req.RemoteAddr),
				// slog.String("userAgent", r.UserAgent()),
				slog.String("referer", req.Referer()),
			}

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			end := time.Now()

			respAttributes := []slog.Attr{
				slog.Time("time", end),
				slog.Duration("latency", end.Sub(start)),
				slog.Int("status", res.Status),
			}

			attributes := []slog.Attr{
				{
					Key:   "request",
					Value: slog.GroupValue(reqAttr...),
				},
				{
					Key:   "response",
					Value: slog.GroupValue(respAttributes...),
				},
			}

			logger.LogAttrs(c.Request().Context(), slog.LevelInfo, "Incoming request", attributes...)

			return nil
		}
	}
}
