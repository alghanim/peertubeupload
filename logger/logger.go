package logger

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type CustomFormatter struct {
	log.TextFormatter
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&CustomFormatter{
		log.TextFormatter{
			FullTimestamp: true,
		},
	})
}

func (f *CustomFormatter) Format(entry *log.Entry) ([]byte, error) {
	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	b := &strings.Builder{}

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = time.RFC3339
	}

	levelColor := f.getColor(entry.Level)
	levelText := strings.ToUpper(entry.Level.String())

	fmt.Fprintf(b, "\x1b[%dm%s\x1b[0m[%s] %s ", levelColor, levelText, entry.Time.Format(timestampFormat), entry.Message)

	for _, k := range keys {
		v := entry.Data[k]
		fmt.Fprintf(b, " %s=%+v", k, v)
	}

	b.WriteByte('\n')
	return []byte(b.String()), nil
}

func (f *CustomFormatter) getColor(level log.Level) int {
	switch level {
	case log.DebugLevel, log.TraceLevel:
		return 37 // white
	case log.InfoLevel:
		return 36 // cyan
	case log.WarnLevel:
		return 33 // yellow
	case log.ErrorLevel, log.FatalLevel, log.PanicLevel:
		return 31 // red
	default:
		return 37 // white
	}
}

func LogInfo(message string, fields map[string]interface{}) {
	if fields != nil {
		log.WithFields(fields).Info(message)
	} else {
		log.Info(message)
	}
}

func LogWarning(message string, fields map[string]interface{}) {
	if fields != nil {
		log.WithFields(fields).Warn(message)
	} else {
		log.Warn(message)
	}
}

func LogError(message string, fields map[string]interface{}) {
	if fields != nil {
		log.WithFields(fields).Error(message)
	} else {
		log.Error(message)
	}
}
