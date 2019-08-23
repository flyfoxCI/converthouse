package util

import (
	"github.com/fagongzi/log"
	"github.com/infinivision/prophet"
)

// InitProphetLog set prophet logger
func InitProphetLog() {
	prophet.SetLogger(&adapterLog{})
}

type adapterLog struct{}

func (l *adapterLog) Info(v ...interface{}) {
	log.Info(v...)
}

func (l *adapterLog) Infof(format string, v ...interface{}) {
	log.Infof(format, v...)
}

func (l *adapterLog) Debug(v ...interface{}) {
	log.Debug(v...)
}

func (l *adapterLog) Debugf(format string, v ...interface{}) {
	log.Debugf(format, v...)
}

func (l *adapterLog) Warn(v ...interface{}) {
	log.Warn(v...)
}

func (l *adapterLog) Warnf(format string, v ...interface{}) {
	log.Warnf(format, v...)
}

func (l *adapterLog) Error(v ...interface{}) {}

func (l *adapterLog) Errorf(format string, v ...interface{}) {
	log.Errorf(format, v...)
}

func (l *adapterLog) Fatal(v ...interface{}) {
	log.Fatal(v...)
}

func (l *adapterLog) Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}
