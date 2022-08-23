package config

import (
	"fmt"
	"os"

	"github.com/RichardKnop/machinery/v1/log"
	"github.com/rs/zerolog"
)

func SetupMachineryLogging(logger *zerolog.Logger) {
	log.SetDebug(&levelLogger{logger, zerolog.DebugLevel})
	log.SetInfo(&levelLogger{logger, zerolog.InfoLevel})
	log.SetWarning(&levelLogger{logger, zerolog.WarnLevel})
	log.SetError(&levelLogger{logger, zerolog.ErrorLevel})
	log.SetFatal(&levelLogger{logger, zerolog.FatalLevel})
}

type levelLogger struct {
	logger *zerolog.Logger
	level  zerolog.Level
}

func (l *levelLogger) event() *zerolog.Event {
	return l.logger.WithLevel(l.level)
}

func (l *levelLogger) Print(v ...interface{}) {
	l.event().Msg(fmt.Sprint(v...))
}

func (l *levelLogger) Printf(format string, v ...interface{}) {
	l.event().Msgf(format, v...)
}

func (l *levelLogger) Println(v ...interface{}) {
	// We don't need to worry about newlines
	l.Print(v...)
}

func (l *levelLogger) Fatal(v ...interface{}) {
	l.event().Msg(fmt.Sprint(v...))
	os.Exit(1)
}

func (l *levelLogger) Fatalf(format string, v ...interface{}) {
	l.event().Msgf(format, v...)
	os.Exit(1)
}

func (l *levelLogger) Fatalln(v ...interface{}) {
	l.Fatal(v...)
}

func (l *levelLogger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.event().Msg(s)
	panic(s)
}

func (l *levelLogger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	l.event().Msg(s)
	panic(s)
}

func (l *levelLogger) Panicln(v ...interface{}) {
	l.Panic(v...)
}
