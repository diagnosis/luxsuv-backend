package logger

import "log"

type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		infoLogger:  log.New(log.Writer(), "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(log.Writer(), "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *Logger) Info(msg string) {
	l.infoLogger.Println(msg)
}

func (l *Logger) Error(msg string) {
	l.errorLogger.Println(msg)
}
