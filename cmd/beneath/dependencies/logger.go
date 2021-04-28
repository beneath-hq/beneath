package dependencies

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/beneath-hq/beneath/cmd/beneath/cli"
	"github.com/beneath-hq/beneath/pkg/envutil"
)

func init() {
	cli.AddDependency(initLogger)
}

func initLogger() *zap.Logger {
	env := envutil.GetEnv()
	isGKE := os.Getenv("IS_GKE")

	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})

	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel
	})

	consoleDebugging := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)

	var encoder zapcore.Encoder
	if env == envutil.Production {
		encoder = zapcore.NewJSONEncoder(newProductionEncoderConfig())
	} else {
		encoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	}

	errorsCore := zapcore.NewCore(encoder, consoleErrors, highPriority)
	if isGKE != "" {
		errorsCore = addGCPErrorLabel(errorsCore)
	}
	debuggingCore := zapcore.NewCore(encoder, consoleDebugging, lowPriority)
	core := zapcore.NewTee(errorsCore, debuggingCore)

	logger := zap.New(
		core,
		zap.AddStacktrace(zap.ErrorLevel),
		zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSampler(core, time.Second, int(100), int(100))
		}),
	)

	return logger
}

// Customize the Production EncoderConfig to use "message" field name
func newProductionEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.EpochTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func addGCPErrorLabel(core zapcore.Core) zapcore.Core {
	gcpErrorLabel := zapcore.Field{
		Key:    "@type",
		Type:   zapcore.StringType,
		String: "type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent",
	}
	return core.With([]zapcore.Field{gcpErrorLabel})
}
