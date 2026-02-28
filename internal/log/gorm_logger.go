package log

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"gorm.io/gorm/logger"
)

const (
	colorReset  = "\033[0m"
	colorPurple = 35
	colorGray   = 90
)

func gormColorize(colorCode int, v string) string {
	return fmt.Sprintf("\033[%sm%s%s", strconv.Itoa(colorCode), v, colorReset)
}

// GormLogger adapts slog for GORM with environment-based formatting.
type GormLogger struct {
	logger *slog.Logger
	env    string
	level  logger.LogLevel
}

// NewGormLogger creates a GORM logger that logs concisely in dev and verbosely in prod.
func NewGormLogger(slogger *slog.Logger, env string) *GormLogger {
	return &GormLogger{
		logger: slogger.With("component", "gorm"),
		env:    env,
		level:  logger.Info,
	}
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return &GormLogger{
		logger: l.logger,
		env:    l.env,
		level:  level,
	}
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...any) {
	if l.level >= logger.Info {
		l.logger.InfoContext(ctx, fmt.Sprintf(msg, data...))
	}
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...any) {
	if l.level >= logger.Warn {
		l.logger.WarnContext(ctx, fmt.Sprintf(msg, data...))
	}
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...any) {
	if l.level >= logger.Error {
		l.logger.ErrorContext(ctx, fmt.Sprintf(msg, data...))
	}
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	if l.env == "prod" { //nolint:nestif
		// Verbose structured logging for production
		attrs := []slog.Attr{
			slog.String("sql", sql),
			slog.Int64("rows", rows),
			slog.Int64("elapsedMs", elapsed.Milliseconds()),
		}
		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
			l.logger.LogAttrs(ctx, slog.LevelError, "query", attrs...)
		} else {
			l.logger.LogAttrs(ctx, slog.LevelDebug, "query", attrs...)
		}
	} else {
		// Concise logging for development
		const maxSQLLen = 120
		truncated := truncateSQL(sql, maxSQLLen)
		if err != nil {
			l.logger.ErrorContext(ctx, fmt.Sprintf("%s %s %s | %s",
				gormColorize(colorPurple, truncated),
				gormColorize(colorGray, fmt.Sprintf("[%d rows]", rows)),
				elapsed.Round(time.Millisecond),
				gormColorize(colorGray, err.Error())))
		} else {
			l.logger.DebugContext(ctx, fmt.Sprintf("%s %s %s",
				gormColorize(colorPurple, truncated),
				gormColorize(colorGray, fmt.Sprintf("[%d rows]", rows)),
				elapsed.Round(time.Millisecond)))
		}
	}
}

func truncateSQL(sql string, maxLen int) string {
	if len(sql) <= maxLen {
		return sql
	}
	return sql[:maxLen] + "..."
}
