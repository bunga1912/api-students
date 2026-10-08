package config

import (
	"log"
	"os"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

func InitLogger() {
	InfoLogger = log.New(
		os.Stdout,
		"[INFO] ",
		log.Ldate|log.Ltime|log.Lshortfile,
	)

	ErrorLogger = log.New(
		os.Stderr,
		"[ERROR] ",
		log.Ldate|log.Ltime|log.Lshortfile,
	)
}

func Info(message string) {
	if InfoLogger == nil {
		InitLogger()
	}

	InfoLogger.Println(message)
}

func Error(message string) {
	if ErrorLogger == nil {
		InitLogger()
	}

	ErrorLogger.Println(message)
}
