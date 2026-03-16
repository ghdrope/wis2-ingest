package utils

import (
	"fmt"
	"os"

	"github.com/akuity/kargo/pkg/logging"
	kargo_os "github.com/akuity/kargo/pkg/os"
)

// GetLogVars resolves logging configuration from environment variables.
//
// This function is taken as is from the Kargo project. In the original Kargo
// repository, this logic exists inside the main package and therefore
// cannot be imported as a reusable library. For this reason, the implementation
// has been copied here to preserve identical logging behavior.
//
// The following environment variables are supported:
//
//	LOG_LEVEL   – Logging level (debug, info, warn, error)
//	LOG_FORMAT  – Log output format (json, console)
//
// If invalid values are provided, safe defaults are applied.
func GetLogVars() (logging.Level, logging.Format) {
	logLevelStr := kargo_os.GetEnv(logging.LogLevelEnvVar, "info")
	logLevel, err := logging.ParseLevel(logLevelStr)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"invalid LOG_LEVEL %q, defaulting to info: %v\n",
			logLevelStr,
			err,
		)
		logLevel = logging.InfoLevel
	}

	logFormatStr := kargo_os.GetEnv(logging.LogFormatEnvVar, string(logging.DefaultFormat))
	logFormat, err := logging.ParseFormat(logFormatStr)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"invalid LOG_FORMAT %q, defaulting to %q: %v\n",
			logFormatStr,
			logging.DefaultFormat,
			err,
		)
		logFormat = logging.DefaultFormat
	}

	return logLevel, logFormat
}
