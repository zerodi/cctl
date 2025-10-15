package logx

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Configure configures zerolog with the desired format ("console"|"json") and debug level.
func Configure(format string, debug bool) {
	// Level
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Output format
	switch strings.ToLower(format) {
	case "json":
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	default: // console by default for a friendlier CLI
		cw := zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
			w.Out = os.Stdout
			w.TimeFormat = time.RFC3339
		})
		log.Logger = zerolog.New(cw).With().Timestamp().Logger()
	}
}
