package log

import (
	"flag"

	"github.com/go-logr/logr"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logger logr.Logger

func GetLogger() logr.Logger {
	if logger == nil {
		if flag.Lookup("development").Value.String() == "true" {
			logger = logf.ZapLogger(true)
		} else {
			logger = logf.ZapLogger(false)
		}
	}
	return logger
}
