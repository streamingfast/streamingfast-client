package cmd

import (
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
)

var traceEnabled = logging.IsTraceEnabled("sf", "github.com/streamingfast/streamingfast-client")

var zlog = zap.NewNop()

func init() {
	logging.Register("github.com/streamingfast/streamingfast-client/cmd", &zlog)
}
