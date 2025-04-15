package glog

import "github.com/sirupsen/logrus"

type DefaultFormatter struct {
	logrus.Formatter
}

func (u DefaultFormatter) Format(e *logrus.Entry) ([]byte, error) {
	e.Time = e.Time.UTC()
	return u.Formatter.Format(e)
}
