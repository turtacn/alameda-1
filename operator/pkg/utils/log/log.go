package log

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/go-logr/logr"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logger logr.Logger

func GetLogger() logr.Logger {
	if logger != nil {
		return logger
	}
	isDev := (flag.Lookup("development").Value.String() == "true")
	dateStr := time.Now().Format(time.RFC3339)
	dateStr = strings.Replace(dateStr, ":", "-", -1)
	f, err := os.Create("operator-" + dateStr + ".log")
	if err != nil {
		logger = logf.ZapLogger(isDev)
	} else {
		logger = logf.ZapLoggerTo(f, isDev)
	}

	return logger
}
