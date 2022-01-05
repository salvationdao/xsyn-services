package log_helpers

import (
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
)

const (
	DEVELOPMENT string = "development"
	TESTING     string = "testing"
	TRAINING    string = "training"
	STAGING     string = "staging"
	PRODUCTION  string = "production"
)

var (
	ErrSentryInitEnvironment = fmt.Errorf("sentry init skipped: invalid environment should be one of %v", []string{DEVELOPMENT, TESTING, TRAINING, STAGING, PRODUCTION})
	ErrSentryInitDSN         = fmt.Errorf("sentry init skipped: dsn missing")
	ErrSentryInitVersion     = fmt.Errorf("sentry init skipped: version missing")
)

func SentryInit(sentryDSNBackend, sentryServerName, version string, sentryEnvironment string, sentryTraceRate float64, log *zerolog.Logger) error {
	switch sentryEnvironment {
	case DEVELOPMENT, TESTING, TRAINING, STAGING, PRODUCTION:
		break
	default:
		return terror.Panic(ErrSentryInitEnvironment, "got", sentryEnvironment)
	}

	if len(sentryDSNBackend) == 0 {
		if sentryEnvironment == PRODUCTION {
			return terror.Panic(ErrSentryInitDSN)
		}
		log.Warn().Err(ErrSentryInitDSN).Msg("")
		return nil
	}
	if len(version) == 0 {
		if sentryEnvironment == PRODUCTION {
			return terror.Panic(ErrSentryInitVersion)
		}
		log.Warn().Err(ErrSentryInitVersion).Msg("")
		return nil
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              sentryDSNBackend,
		ServerName:       sentryServerName,
		Environment:      string(sentryEnvironment),
		Release:          version,
		TracesSampleRate: sentryTraceRate,
		AttachStacktrace: false, // Errors will be merged based on stack trace.
		// Unfortunately terror interfears with the stack trace that stdlib.Error collects
		// causing all sentry stack traces to look alike. Creating matching errors
	})
	if err != nil {
		return terror.Error(fmt.Errorf("sentry init failed: %v", err))
	}
	log.Info().Msg("Sentry Initialised")
	return nil
}

func LogToSentry(sentryHub *sentry.Hub, level sentry.Level, msg string, echo string) {
	sentryHub.WithScope(func(scope *sentry.Scope) {
		// configure message
		scope.SetLevel(level)
		scope.SetExtra("friendly_message", msg)
		scope.SetExtra("echo", echo)

		sentry.CaptureMessage(msg)
	})
}
