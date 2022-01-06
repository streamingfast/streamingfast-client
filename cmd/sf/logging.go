package main

import (
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
)

var zlog *zap.Logger

func init() {
	logging.ApplicationLogger("sf", "github.com/streamingfast/streamingfast-client", &zlog)
}
