package conf

import (
	"github.com/sirupsen/logrus"
	"fmt"
	"os"
	"gopkg.in/natefinch/lumberjack.v2"
)

// A hook that prints logs with level Warn and up always to the stderr, even if a log file is written.
type StdErrLogHook struct {
}

func (h *StdErrLogHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.WarnLevel, logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel}
}
func (h *StdErrLogHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read entry, %v", err)
		return err
	}
	fmt.Fprintf(os.Stderr, line)
	return nil
}

func SetupLogging(config MiscConf) {
	logrus.SetFormatter(&logrus.TextFormatter{DisableColors: false, DisableSorting: true, FullTimestamp: true, ForceColors:true})
	if config.DebugLogging {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	if config.Logfile == "" {
		logrus.SetOutput(os.Stdout)
	} else {
		logrus.SetOutput(
			&lumberjack.Logger{
				Filename:   config.Logfile,
				MaxSize:    20, // megabytes
				MaxBackups: 3,
				MaxAge:     90,   //days
				Compress:   true, // disabled by default
			})

		logrus.AddHook(&StdErrLogHook{})
	}
}
