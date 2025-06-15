package log

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

type cronLogger struct {
	logger *slog.Logger
}

func (l *cronLogger) Info(msg string, keysAndValues ...any) {
	keysAndValues = formatTimes(keysAndValues)
	l.logger.Debug(fmt.Sprintf("cron: %s", msg), keysAndValues...)
}

func (l *cronLogger) Error(err error, msg string, keysAndValues ...any) {
	keysAndValues = formatTimes(keysAndValues)
	l.logger.Error(fmt.Sprintf("cron: %s", msg), keysAndValues...)
}

// formatTimes formats any time.Time values as RFC3339.
// from github.com/robfig/cron/v3
func formatTimes(keysAndValues []any) []any {
	var formattedArgs []any
	for _, arg := range keysAndValues {
		if t, ok := arg.(time.Time); ok {
			arg = t.Format(time.RFC3339)
		}
		formattedArgs = append(formattedArgs, arg)
	}
	return formattedArgs
}

func NewCronLogger(logger *slog.Logger) cron.Logger {
	return &cronLogger{logger}
}
