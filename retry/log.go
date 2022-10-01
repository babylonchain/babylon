package retry

import (
	"github.com/sirupsen/logrus"
	"os"
)

var logger = &logrus.Logger{
	Out:   os.Stderr,
	Level: logrus.DebugLevel,
	Formatter: &logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	},
}

var log = logger.WithField("module", "retry")
