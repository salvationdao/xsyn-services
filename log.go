package passport

import (
	"passport/canlog"
	"context"
	
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// WithCanonicalLogger used to construct a new contextual logger to log one line per request
// This helps reduce log bloat https://stripe.com/au/blog/canonical-log-lines and increase observability
func WithCanonicalLogger(ctx context.Context, zlog *zap.Logger) context.Context {
	canlog.DefaultLogger = &canlog.ZapLogger{Logger: zlog, DetectErr: true}
	newctx := canlog.NewContext(ctx)
	return newctx
}

// NewLogToSyslog returns new syslogger for Zap

// NewLogToFile creates a new file logger
func NewLogToFile(filename, tag, version string, prod bool) *zap.SugaredLogger {
	var conf zap.Config
	if !prod {
		conf = zap.NewDevelopmentConfig()
		conf.OutputPaths = []string{
			filename,
		}
	} else {
		conf = zap.NewProductionConfig()
		conf.OutputPaths = []string{
			filename,
		}
	}

	l, err := conf.Build(
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zap.FatalLevel),
	)
	if err != nil {
		panic(err)
	}
	return l.Sugar().With("version", version).With("tag", tag)
}

// NewLogToStdOut creates a new file logger
func NewLogToStdOut(tag, version string, prod bool) *zap.SugaredLogger {
	if prod {
		config := zap.NewProductionConfig()
		l, err := config.Build()
		if err != nil {
			panic("can't initialize zap logger: " + err.Error())
		}
		return l.Sugar().With("version", version).With("tag", tag)
	}

	config := zap.NewDevelopmentConfig()
	config.DisableStacktrace = true
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	l, err := config.Build()
	if err != nil {
		panic("can't initialize zap logger: " + err.Error())
	}
	return l.Sugar().With("version", version).With("tag", tag)
}
