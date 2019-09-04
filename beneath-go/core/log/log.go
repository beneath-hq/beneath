package log

import (
	"go.uber.org/zap"
)

var (
	// L is the Zap logger
	L *zap.Logger

	// S is a suggared Zap logger
	S *zap.SugaredLogger
)

func init() {
	L, _ = zap.NewProduction()
	S = L.Sugar()
}
