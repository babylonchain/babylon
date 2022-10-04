package retry

import (
	"github.com/tendermint/tendermint/libs/log"
	"os"
)

// TODO add log formatters
var logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
