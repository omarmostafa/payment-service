package logger

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const LogFormatFriendly = "friendly"

type Debug struct {
	Id      string
	Enabled bool
	Level   string
	Format  string
}

type Logger struct {
	Config     *Debug
	logService *zap.SugaredLogger
}

func GetLogger() *Logger {
	debugConfig := &Debug{
		Id:      "payment-service",
		Enabled: true,
		Level:   "debug",
		Format:  "friendly",
	}
	logger := &Logger{Config: debugConfig}
	_ = logger.Setup()
	return logger
}

// Debug uses fmt.Sprint to construct and log a message.
func (l *Logger) Debug(args ...interface{}) {
	l.logService.Debug(args)
}

// Info uses fmt.Sprint to construct and log a message.
func (l *Logger) Info(args ...interface{}) {
	l.logService.Info(args...)
}

// Warn uses fmt.Sprint to construct and log a message.
func (l *Logger) Warn(args ...interface{}) {
	l.logService.Warn(args...)
}

// Error uses fmt.Sprint to construct and log a message.
func (l *Logger) Error(args ...interface{}) {
	l.logService.Error(args...)
}

// Error uses fmt.Sprint to construct and log a message.
func (l *Logger) Panic(args ...interface{}) {
	l.logService.Panic(args...)
}

// Error uses fmt.Sprint to construct and log a message.
func (l *Logger) Fatal(args ...interface{}) {
	l.logService.Fatal(args)
}

// Debugf uses fmt.Sprintf to log a templated message.
func (l *Logger) Debugf(template string, args ...interface{}) {
	l.logService.Debugf(template, args...)
}

// Infof uses fmt.Sprintf to log a templated message.
func (l *Logger) Infof(template string, args ...interface{}) {
	l.logService.Infof(template, args...)
}

// Warnf uses fmt.Sprintf to log a templated message.
func (l *Logger) Warnf(template string, args ...interface{}) {
	l.logService.Warnf(template, args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func (l *Logger) Errorf(template string, args ...interface{}) {
	l.logService.Errorf(template, args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func (l *Logger) Panicf(template string, args ...interface{}) {
	l.logService.Panicf(template, args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.logService.Fatalf(template, args...)
}

func (l *Logger) setupFriendlyLogger() zap.Config {
	encodeConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     "\n\n",
		EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}

	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(l.GetLogLevel()),
		Development: true,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		InitialFields:    map[string]interface{}{},
		Encoding:         "console",
		EncoderConfig:    encodeConfig,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
	return cfg
}

func (l *Logger) Setup() error {

	cfg := zap.Config{}
	cfg = l.setupFriendlyLogger()

	logger, err := cfg.Build()
	if err != nil {
		log.Panic(err)
		return err
	}

	defer logger.Sync()

	l.logService = logger.Sugar()
	return nil
}

func (l *Logger) GetLogLevel() zapcore.Level {
	// set default to warn
	logLevel := zap.WarnLevel

	// if logs disabled, set logLevel to error
	if !l.IsLogEnabled() {
		logLevel = zap.ErrorLevel
	}

	configLevel := strings.ToLower(l.Config.Level)

	if configLevel != "" {
		switch configLevel {
		case "debug":
			logLevel = zap.DebugLevel
			break
		case "info":
			logLevel = zap.InfoLevel
			break
		case "warn":
			logLevel = zap.WarnLevel
			break
		case "error":
			logLevel = zap.ErrorLevel
			break
		}
	}

	return logLevel
}

func (l *Logger) IsLogEnabled() bool {
	if l.Config == nil {
		return true
	}

	return l.Config.Enabled
}

func NewStructuredLogger(logger *logrus.Logger) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&StructuredLogger{logger})
}

// @todo replace logrus with zap
type StructuredLogger struct {
	Logger *logrus.Logger
}

func (l *StructuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	entry := &StructuredLoggerEntry{Logger: logrus.NewEntry(l.Logger)}
	logFields := logrus.Fields{}

	logFields["ts"] = time.Now().UTC().Format(time.RFC1123)

	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		logFields["req_id"] = reqID
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	logFields["http_scheme"] = scheme
	logFields["http_proto"] = r.Proto
	logFields["http_method"] = r.Method

	logFields["remote_addr"] = r.RemoteAddr
	logFields["user_agent"] = r.UserAgent()

	logFields["uri"] = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)

	entry.Logger = entry.Logger.WithFields(logFields)

	entry.Logger.Infoln("request started")

	return entry
}

type StructuredLoggerEntry struct {
	Logger logrus.FieldLogger
}

func (l *StructuredLoggerEntry) Write(status, bytes int, elapsed time.Duration) {
	l.Logger = l.Logger.WithFields(logrus.Fields{
		"resp_status": status, "resp_bytes_length": bytes,
		"resp_elapsed_ms": float64(elapsed.Nanoseconds()) / 1000000.0,
	})

	l.Logger.Infoln("request complete")
}

func (l *StructuredLoggerEntry) Panic(v interface{}, stack []byte) {
	l.Logger = l.Logger.WithFields(logrus.Fields{
		"stack": string(stack),
		"panic": fmt.Sprintf("%+v", v),
	})
}
