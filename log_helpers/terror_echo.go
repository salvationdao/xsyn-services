package log_helpers

import (
	"errors"
	stdlog "log"

	"github.com/getsentry/sentry-go"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
)

func TerrorEcho(sentryHub *sentry.Hub, err error, log *zerolog.Logger) {
	stdlog.SetPrefix("TERROR: ")
	lvl := terror.GetLevel(err)
	msg := err.Error()
	var bErr *terror.TError
	if errors.As(err, &bErr) {
		msg = bErr.Message
	}

	echo := terror.Echo(err, true)

	switch lvl {
	case terror.ErrLevelPanic:
		LogToSentry(sentryHub, sentry.LevelFatal, msg, echo)
		// using WithLevel to prevent zerolog from calling `panic()`
		log.WithLevel(zerolog.PanicLevel).Caller(1).Err(err).Msg(msg)
		terror.Echo(err)
	case terror.ErrLevelError:
		LogToSentry(sentryHub, sentry.LevelError, msg, echo)
		log.Error().Caller(1).Err(err).Msg(msg)
		terror.Echo(err)
	case terror.ErrLevelWarn:
		LogToSentry(sentryHub, sentry.LevelWarning, msg, echo)
		log.Warn().Caller(1).Err(err).Msg(msg)
	default:
		LogToSentry(sentryHub, sentry.LevelError, msg, echo)
		log.Info().Caller(1).Err(err).Msg(msg)
	}
}
