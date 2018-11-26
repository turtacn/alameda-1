package log

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/go-logr/logr"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// TODO: Deprecate all this stuff

var logger logr.Logger

func GetLogger() logr.Logger {
	if logger != nil {
		return logger
	}

	isDev := false
	if devFlag := flag.Lookup("development"); devFlag != nil {
		isDev = (flag.Lookup("development").Value.String() == "true")
	}

	isLogOutput := false
	if isLogOutputFlag := flag.Lookup("logfile"); isLogOutputFlag != nil {
		isLogOutput = (flag.Lookup("logfile").Value.String() == "true")
	}

	dateStr := time.Now().Format(time.RFC3339)
	dateStr = strings.Replace(dateStr, ":", "-", -1)

	if !isLogOutput {
		logger = logf.ZapLogger(isDev)
	} else if f, err := os.Create("operator-" + dateStr + ".log"); err == nil {
		logger = logf.ZapLoggerTo(f, isDev)
	} else {
		logger = logf.ZapLogger(isDev)
	}

	return logger
}
