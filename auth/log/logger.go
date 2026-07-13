package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var globalLogger *zap.Logger

func InitLogger(env string, level string) (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	lvl, err := zapcore.ParseLevel(level)
	if err != nil {
		lvl = zapcore.InfoLevel
	}

	if env == "dev" {
		cfg := zap.NewDevelopmentConfig()
		cfg.Level = zap.NewAtomicLevelAt(lvl)
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger, err = cfg.Build()
	} else {
		cfg := zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(lvl)

		_ = os.MkdirAll("logs", 0755)

		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   "logs/app.log",
			MaxSize:    50,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   true,
		})
		consoleWriter := zapcore.AddSync(os.Stdout)
		multiWriter := zapcore.NewMultiWriteSyncer(fileWriter, consoleWriter)

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(cfg.EncoderConfig),
			multiWriter,
			cfg.Level,
		)
		logger = zap.New(core)
	}

	if err != nil {
		return nil, err
	}

	globalLogger = logger
	return logger, nil
}

func GetLogger() *zap.Logger {
	if globalLogger == nil {
		l, _ := zap.NewDevelopment()
		globalLogger = l
	}
	return globalLogger
}

func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

func SetGlobalLogger(logger *zap.Logger) {
	globalLogger = logger
}
