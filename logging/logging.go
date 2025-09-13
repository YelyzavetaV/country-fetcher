package logging

import (
	"os"
	"fmt"
	"log/slog"
)

var logLevelMap = map[string]slog.Level{
	"DEBUG": slog.LevelDebug,
	"INFO":  slog.LevelInfo,
	"WARN":  slog.LevelWarn,
	"ERROR": slog.LevelError,
}

var Log *slog.Logger

func InitLogger(opts slog.HandlerOptions) {
	Log = slog.New(slog.NewJSONHandler(os.Stdout, &opts))
}

func ParseLogLevel(s string) (slog.Level, error) {
	if lvl, ok := logLevelMap[s]; ok {
		return lvl, nil
	}
	return slog.LevelWarn, fmt.Errorf(
		"Could not parse log level: %s; falling back to WARN\n", s)
}