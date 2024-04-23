package main

import (
	"fmt"
	"strings"

	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

func GooseZapLogger(l *zap.Logger) goose.Logger {
	return &gooseZapLogger{l: l}
}

type gooseZapLogger struct {
	l *zap.Logger
}

func (g *gooseZapLogger) Printf(format string, v ...interface{}) {
	g.l.Info(fmt.Sprintf(strings.Replace(format, "\n", "", 1), v...))
}

func (g *gooseZapLogger) Fatalf(format string, v ...interface{}) {
	g.l.Fatal(fmt.Sprintf(strings.Replace(format, "\n", "", 1), v...))
}
