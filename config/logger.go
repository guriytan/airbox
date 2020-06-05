package config

import (
	"log"
	"os"
)

var logger *log.Logger

func InitializeLogger() {
	logger = log.New(os.Stderr, "AirBox", log.LstdFlags|log.Lshortfile)
}

func GetLogger() *log.Logger {
	return logger
}
