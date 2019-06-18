package gateway

import (
	"github.com/beneath-core/beneath-gateway/beneath"
)

var (
	engine *beneath.Engine
)

func init() {
	engine = beneath.NewEngine()
}
