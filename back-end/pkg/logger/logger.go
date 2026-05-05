package logger

import (
	"fmt"
	"mm/config"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const timeFormat = "2006-01-02 15:04:05"

func InitLogger(c *config.Config) *zap.Logger {
	level, err := zapcore.ParseLevel(c.App.LogLevel)
	if err != nil {
		level = zap.InfoLevel
	}

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:     "time",
		LevelKey:    "level",
		NameKey:     "logger",
		MessageKey:  "msg",
		EncodeLevel: zapcore.CapitalLevelEncoder,
		EncodeTime:  zapcore.TimeEncoderOfLayout(timeFormat),
	}

	var encoder zapcore.Encoder

	switch c.App.Environment {
	case config.EnvironmentProduction, config.EnvironmentStage:
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	default:
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		zap.NewAtomicLevelAt(level),
	)

	logger := zap.New(core)
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Println("failed to sync logger:", err)
		}
	}()

	cfg := zap.NewProductionConfig()
	zap.Must(cfg.Build())

	return logger
}
