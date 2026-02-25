package monitoring

import (
	"log"
	"os"
)

var logger *log.Logger
var logFile *os.File

// SetupLogger sets up file logging if path != ""
func SetupLogger(path string) error {
	if path == "" {
		// Use default stdout logger
		logger = log.Default()
		return nil
	}

	var err error
	logFile, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	logger = log.New(logFile, "", log.LstdFlags|log.Lshortfile)
	log.SetOutput(logger.Writer())
	return nil
}

func ShutdownLogger() error {
	if logFile != nil {
		return logFile.Close()
	}
	return nil
}

func GetLogger() *log.Logger {
	return logger
}
