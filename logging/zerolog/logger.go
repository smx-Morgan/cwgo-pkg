// Copyright 2024 CloudWeGo Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package zerolog

import (
	"context"
	"fmt"
	"github.com/kitex-contrib/obs-opentelemetry/logging"
	"io"
	"os"

	"github.com/rs/zerolog"
)

var _ logging.FullLogger = (*Logger)(nil)

// Ref to https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/README.md#json-formats
const (
	traceIDKey    = "trace_id"
	spanIDKey     = "span_id"
	traceFlagsKey = "trace_flags"
)

type Logger struct {
	l      *zerolog.Logger
	config *config
}

func NewLogger(opts ...Option) *Logger {
	cfg := defaultConfig()

	// apply options
	for _, opt := range opts {
		opt.apply(cfg)
	}

	// default logger
	logger := zerolog.New(os.Stdout)
	if cfg.logger != nil {
		logger = *cfg.logger
	}

	zerologLogger := logger.Hook(cfg.defaultZerologHookFn())

	return &Logger{
		l:      &zerologLogger,
		config: cfg,
	}
}

func (l *Logger) Logger() *zerolog.Logger {
	return l.l
}

// Log log using zerolog logger with specified level
func (l *Logger) Log(level logging.Level, kvs ...any) {
	switch level {
	case logging.LevelTrace, logging.LevelDebug:
		l.l.Debug().Msg(fmt.Sprint(kvs...))
	case logging.LevelInfo:
		l.l.Info().Msg(fmt.Sprint(kvs...))
	case logging.LevelNotice, logging.LevelWarn:
		l.l.Warn().Msg(fmt.Sprint(kvs...))
	case logging.LevelError:
		l.l.Error().Msg(fmt.Sprint(kvs...))
	case logging.LevelFatal:
		l.l.Fatal().Msg(fmt.Sprint(kvs...))
	default:
		l.l.Warn().Msg(fmt.Sprint(kvs...))
	}
}

// Logf log using zerolog logger with specified level and formatting
func (l *Logger) Logf(level logging.Level, format string, kvs ...any) {
	switch level {
	case logging.LevelTrace, logging.LevelDebug:
		l.l.Debug().Msg(fmt.Sprintf(format, kvs...))
	case logging.LevelInfo:
		l.l.Info().Msg(fmt.Sprintf(format, kvs...))
	case logging.LevelNotice, logging.LevelWarn:
		l.l.Warn().Msg(fmt.Sprintf(format, kvs...))
	case logging.LevelError:
		l.l.Error().Msg(fmt.Sprintf(format, kvs...))
	case logging.LevelFatal:
		l.l.Fatal().Msg(fmt.Sprintf(format, kvs...))
	default:
		l.l.Warn().Msg(fmt.Sprintf(format, kvs...))
	}
}

// CtxLogf log with logger associated with context.
// If no logger is associated, DefaultContextLogger is used, unless DefaultContextLogger is nil, in which case a disabled logger is used.
func (l *Logger) CtxLogf(level logging.Level, ctx context.Context, format string, kvs ...any) {
	logger := l.Logger()
	// todo add hook
	switch level {
	case logging.LevelTrace, logging.LevelDebug:
		logger.Debug().Ctx(ctx).Msg(fmt.Sprintf(format, kvs...))
	case logging.LevelInfo:
		logger.Info().Ctx(ctx).Msg(fmt.Sprintf(format, kvs...))
	case logging.LevelNotice, logging.LevelWarn:
		logger.Warn().Ctx(ctx).Msg(fmt.Sprintf(format, kvs...))
	case logging.LevelError:
		logger.Error().Ctx(ctx).Msg(fmt.Sprintf(format, kvs...))
	case logging.LevelFatal:
		logger.Fatal().Ctx(ctx).Msg(fmt.Sprintf(format, kvs...))
	default:
		logger.Warn().Ctx(ctx).Msg(fmt.Sprintf(format, kvs...))
	}
}

// Trace logs a message at trace level.
func (l *Logger) Trace(v ...any) {
	l.Log(logging.LevelTrace, v...)
}

// Debug logs a message at debug level.
func (l *Logger) Debug(v ...any) {
	l.Log(logging.LevelDebug, v...)
}

// Info logs a message at info level.
func (l *Logger) Info(v ...any) {
	l.Log(logging.LevelInfo, v...)
}

// Notice logs a message at notice level.
func (l *Logger) Notice(v ...any) {
	l.Log(logging.LevelNotice, v...)
}

// Warn logs a message at warn level.
func (l *Logger) Warn(v ...any) {
	l.Log(logging.LevelWarn, v...)
}

// Error logs a message at error level.
func (l *Logger) Error(v ...any) {
	l.Log(logging.LevelError, v...)
}

// Fatal logs a message at fatal level.
func (l *Logger) Fatal(v ...any) {
	l.Log(logging.LevelFatal, v...)
}

// Tracef logs a formatted message at trace level.
func (l *Logger) Tracef(format string, v ...any) {
	l.Logf(logging.LevelTrace, format, v...)
}

// Debugf logs a formatted message at debug level.
func (l *Logger) Debugf(format string, v ...any) {
	l.Logf(logging.LevelDebug, format, v...)
}

// Infof logs a formatted message at info level.
func (l *Logger) Infof(format string, v ...any) {
	l.Logf(logging.LevelInfo, format, v...)
}

// Noticef logs a formatted message at notice level.
func (l *Logger) Noticef(format string, v ...any) {
	l.Logf(logging.LevelWarn, format, v...)
}

// Warnf logs a formatted message at warn level.
func (l *Logger) Warnf(format string, v ...any) {
	l.Logf(logging.LevelWarn, format, v...)
}

// Errorf logs a formatted message at error level.
func (l *Logger) Errorf(format string, v ...any) {
	l.Logf(logging.LevelError, format, v...)
}

// Fatalf logs a formatted message at fatal level.
func (l *Logger) Fatalf(format string, v ...any) {
	l.Logf(logging.LevelError, format, v...)
}

// CtxTracef logs a message at trace level with logger associated with context.
// If no logger is associated, DefaultContextLogger is used, unless DefaultContextLogger is nil, in which case a disabled logger is used.
func (l *Logger) CtxTracef(ctx context.Context, format string, v ...any) {
	l.CtxLogf(logging.LevelTrace, ctx, format, v...)
}

// CtxDebugf logs a message at debug level with logger associated with context.
// If no logger is associated, DefaultContextLogger is used, unless DefaultContextLogger is nil, in which case a disabled logger is used.
func (l *Logger) CtxDebugf(ctx context.Context, format string, v ...any) {
	l.CtxLogf(logging.LevelDebug, ctx, format, v...)
}

// CtxInfof logs a message at info level with logger associated with context.
// If no logger is associated, DefaultContextLogger is used, unless DefaultContextLogger is nil, in which case a disabled logger is used.
func (l *Logger) CtxInfof(ctx context.Context, format string, v ...any) {
	l.CtxLogf(logging.LevelInfo, ctx, format, v...)
}

// CtxNoticef logs a message at notice level with logger associated with context.
// If no logger is associated, DefaultContextLogger is used, unless DefaultContextLogger is nil, in which case a disabled logger is used.
func (l *Logger) CtxNoticef(ctx context.Context, format string, v ...any) {
	l.CtxLogf(logging.LevelNotice, ctx, format, v...)
}

// CtxWarnf logs a message at warn level with logger associated with context.
// If no logger is associated, DefaultContextLogger is used, unless DefaultContextLogger is nil, in which case a disabled logger is used.
func (l *Logger) CtxWarnf(ctx context.Context, format string, v ...any) {
	l.CtxLogf(logging.LevelWarn, ctx, format, v...)
}

// CtxErrorf logs a message at error level with logger associated with context.
// If no logger is associated, DefaultContextLogger is used, unless DefaultContextLogger is nil, in which case a disabled logger is used.
func (l *Logger) CtxErrorf(ctx context.Context, format string, v ...any) {
	l.CtxLogf(logging.LevelError, ctx, format, v...)
}

// CtxFatalf logs a message at fatal level with logger associated with context.
// If no logger is associated, DefaultContextLogger is used, unless DefaultContextLogger is nil, in which case a disabled logger is used.
func (l *Logger) CtxFatalf(ctx context.Context, format string, v ...any) {
	l.CtxLogf(logging.LevelFatal, ctx, format, v...)
}

func (l *Logger) SetLevel(level logging.Level) {
	var lv zerolog.Level
	switch level {
	case logging.LevelTrace:
		lv = zerolog.TraceLevel
	case logging.LevelDebug:
		lv = zerolog.DebugLevel
	case logging.LevelInfo:
		lv = zerolog.InfoLevel
	case logging.LevelWarn, logging.LevelNotice:
		lv = zerolog.WarnLevel
	case logging.LevelError:
		lv = zerolog.ErrorLevel
	case logging.LevelFatal:
		lv = zerolog.FatalLevel
	default:
		lv = zerolog.WarnLevel
	}
	l.l.Level(lv)
}

func (l *Logger) SetOutput(writer io.Writer) {
	l.l.Output(writer)
}
