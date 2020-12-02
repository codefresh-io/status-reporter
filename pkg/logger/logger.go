package logger

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

type (
	// Logger to print stuff
	Logger interface {
		Info(msg string, keysAndValues ...interface{})
		V(level int) Logger
		Fork(keysAndValues ...interface{}) Logger
	}

	// Options to build new Logger
	Options struct{}

	log struct {
		lgr logr.Logger
	}
)

// New builds Logger
func New(options Options) Logger {
	zapLog, err := zap.NewDevelopment(zap.AddCallerSkip(1))
	if err != nil {
		panic(fmt.Sprintf("who watches the watchmen (%v)?", err))
	}
	lgr := zapr.NewLogger(zapLog)
	return &log{
		lgr: lgr,
	}
}

func (l *log) Info(msg string, keysAndValues ...interface{}) {
	l.lgr.Info(msg, keysAndValues...)
}
func (l *log) V(level int) Logger {
	return &log{
		lgr: l.lgr.V(level),
	}
}
func (l *log) Fork(keysAndValues ...interface{}) Logger {
	return &log{
		lgr: l.lgr.WithValues(keysAndValues...),
	}
}
