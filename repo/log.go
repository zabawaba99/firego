package repo

import (
	"github.com/Sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func dLog(msg string) {
	logrus.Debugln(msg)
}

func dLogField(msg string, k string, v interface{}) {
	logrus.WithField(k, v).Debugln(msg)
}

func dLogFields(msg string, fields map[string]interface{}) {
	logrus.WithFields(logrus.Fields(fields)).Debugln(msg)
}

func iLogField(msg string, k string, v interface{}) {
	logrus.WithField(k, v).Infoln(msg)
}

func wLogField(msg string, k string, v interface{}) {
	logrus.WithField(k, v).Warningln(msg)
}

func eLog(msg string, err error) {
	logrus.WithError(err).Errorln(msg)
}

func eLogFields(msg string, fields map[string]interface{}) {
	logrus.WithFields(logrus.Fields(fields)).Errorln(msg)
}
