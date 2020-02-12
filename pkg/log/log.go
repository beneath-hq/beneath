package log

import (
	"os"
	"time"

	"github.com/beneath-core/pkg/envutil"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// configSpecification defines the config variables to load from ENV
// See https://github.com/kelseyhightower/envconfig
type configSpecification struct {
	Env string `envconfig:"ENV" required:"false" default:"production"`
}

var (
	// L is the Zap logger
	L *zap.Logger

	// S is a suggared Zap logger
	S *zap.SugaredLogger

	// config parsed from env
	config configSpecification
)

func init() {
	envutil.LoadConfig("", &config)

	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})

	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel
	})

	consoleDebugging := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)

	consoleEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
		zapcore.NewCore(consoleEncoder, consoleDebugging, lowPriority),
	)

	L = zap.New(
		core,
		zap.AddStacktrace(zap.ErrorLevel),
		zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSampler(core, time.Second, int(100), int(100))
		}),
	)
	defer L.Sync()

	S = L.Sugar()
}
