package retry

import (
	"github.com/tendermint/tendermint/libs/log"
	"os"
)

var logger = log.NewTMFmtLogger(log.NewSyncWriter(os.Stdout))
