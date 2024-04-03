package client

import (
	"fmt"
	zaplogfmt "github.com/jsternberg/zap-logfmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

func newRootLogger(format string, debug bool) (*zap.Logger, error) {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = func(ts time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(ts.UTC().Format("2006-01-02T15:04:05.000000Z07:00"))
	}
	cfg.LevelKey = "lvl"

	var enc zapcore.Encoder
	switch format {
	case "json":
		enc = zapcore.NewJSONEncoder(cfg)
	case "auto", "console":
		enc = zapcore.NewConsoleEncoder(cfg)
	case "logfmt":
		enc = zaplogfmt.NewEncoder(cfg)
	default:
		return nil, fmt.Errorf("unrecognized log format %q", format)
	}

	level := zap.InfoLevel
	if debug {
		level = zap.DebugLevel
	}
	return zap.New(zapcore.NewCore(
		enc,
		os.Stderr,
		level,
	)), nil
}
