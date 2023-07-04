package logger

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})
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
