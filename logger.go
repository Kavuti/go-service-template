package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

func NewLogger() *zap.Logger {
	logger := zap.Must(zap.NewProduction())
	if os.Getenv("APP_ENV") != "production" {
		logger = zap.Must(zap.NewDevelopment())
	}
	return logger
}

func NewSugaredLogger() *zap.SugaredLogger {
	logger := zap.Must(zap.NewProduction()).Sugar()
	if os.Getenv("APP_ENV") != "production" {
		logger = zap.Must(zap.NewDevelopment()).Sugar()
	}
	return logger
}

func GooseZapLogger(l *zap.Logger) goose.Logger {
	return &gooseZapLogger{l: l}
}

func GooseZapSugaredLogger(s *zap.SugaredLogger) goose.Logger {
	return &gooseZapSugaredLogger{s: s}
}

type gooseZapLogger struct {
	l *zap.Logger
}

type gooseZapSugaredLogger struct {
	s *zap.SugaredLogger
}

func (g *gooseZapLogger) Printf(format string, v ...interface{}) {
	g.l.Info(fmt.Sprintf(strings.Replace(format, "\n", "", 1), v...))
}

func (g *gooseZapSugaredLogger) Printf(format string, v ...interface{}) {
	g.s.Infof(strings.Replace(format, "\n", "", 1), v...)
}

func (g *gooseZapLogger) Fatalf(format string, v ...interface{}) {
	g.l.Fatal(fmt.Sprintf(strings.Replace(format, "\n", "", 1), v...))
}

func (g *gooseZapSugaredLogger) Fatalf(format string, v ...interface{}) {
	g.s.Fatalf(strings.Replace(format, "\n", "", 1), v...)
}
