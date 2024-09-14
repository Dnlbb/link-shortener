package logger

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

type Logger interface {
	Info(args ...interface{})
	WithFields(fields map[string]interface{}) Logger
}

type LogrusLogger struct {
	entry *logrus.Entry
}

func NewLogrusLogger(logger *logrus.Logger) Logger {
	return &LogrusLogger{entry: logrus.NewEntry(logger)}
}

func (l *LogrusLogger) Info(args ...interface{}) {
	l.entry.Info(args...)
}

func (l *LogrusLogger) WithFields(fields map[string]interface{}) Logger {
	return &LogrusLogger{entry: l.entry.WithFields(logrus.Fields(fields))}
}

type (
	ResponseData struct {
		Status int
		Size   int
	}

	LoggingResponseWriter struct {
		http.ResponseWriter
		ResponseData *ResponseData
	}
)

func (r *LoggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.ResponseData.Size += size
	return size, err
}

func (r *LoggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseData.Status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)

}
