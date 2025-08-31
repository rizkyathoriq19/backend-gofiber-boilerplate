package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

func New(env string) *logrus.Logger {
	logger := logrus.New()
	
	// Set output
	logger.SetOutput(os.Stdout)
	
	// Set log level based on environment
	if env == "production" {
		logger.SetLevel(logrus.InfoLevel)
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetLevel(logrus.DebugLevel)
		logger.SetFormatter(&logrus.TextFormatter{
			ForceColors: true,
		})
	}
	
	return logger
}