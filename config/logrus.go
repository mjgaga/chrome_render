package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"path"
	"runtime"
	"strings"
)

func setLogCof(levelName string) {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02T15:04:05.000Z07",
		CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
			function = f.Function
			if strings.Index(f.File, ProjectName) > -1 {
				ps := strings.Split(f.File, ProjectName)
				file = " ." + ps[1]
			} else {
				file = path.Base(f.File)
			}
			file = fmt.Sprintf("%s:%d", file, f.Line)
			return
		},
	})
	logrus.SetReportCaller(true)
	level, _ := logrus.ParseLevel(levelName)
	logrus.SetLevel(level)
}
