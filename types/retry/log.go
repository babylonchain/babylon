package retry

import (
	"github.com/tendermint/tendermint/libs/log"
	"os"
)

var logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
