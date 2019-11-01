package log

import (
	"github.com/beneath-core/beneath-go/core"

	"go.uber.org/zap"
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
	core.LoadConfig("", &config)

	var zapConfig zap.Config
	if config.Env == "production" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	zapConfig.DisableCaller = true
	zapConfig.OutputPaths = []string{"stdout"}

	L, _ = zapConfig.Build()
	S = L.Sugar()
}
