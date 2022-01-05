package log_helpers

import (
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

func NamedLogger(logger *zerolog.Logger, name string) *zerolog.Logger {
	log := logger.With().Str("name", name).Logger()
	return &log
}

func LoggerInitZero(sentryEnvironment string) *zerolog.Logger {

	output := loggerStdout(sentryEnvironment)

	log := zerolog.New(output).With().Timestamp().Logger()
	return &log
}

func loggerStdout(sentryEnvironment string) zerolog.ConsoleWriter {
	output := zerolog.NewConsoleWriter()

	if sentryEnvironment != "development" {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		output.TimeFormat = time.RFC3339

		output.FormatLevel = func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
		}
		output.FormatMessage = func(i interface{}) string {
			if i == nil {
				return "no msg"
			}
			return fmt.Sprintf("%s", i)
		}
	} else {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}
	return output
}
