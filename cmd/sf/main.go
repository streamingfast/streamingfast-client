package main

import (
	"github.com/streamingfast/logging"
	"github.com/streamingfast/streamingfast-client/cmd/sf/cmd"
)

func init() {
	logging.ApplicationLogger("sf", "github.com/streamingfast/streamingfast-client/cmd/sf")
}

func main() {
	cmd.Execute()
}
