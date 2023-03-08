package retry

import (
	"github.com/cometbft/cometbft/libs/log"
	"os"
)

// TODO add log formatters
var logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
