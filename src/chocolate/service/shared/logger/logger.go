package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

var (
	//_TRACE   *log.Logger
	_info    *log.Logger
	_debug   *log.Logger
	_warning *log.Logger
	_error   *log.Logger
	_logFile *os.File
)

// Init initializes logger
func Init(logPath string, debug bool) error {
	var w io.Writer

	if logPath != "" {
		fmt.Println("Logger init logPath:", logPath)
		var err error
		if _logFile, err = os.Create(logPath); err != nil {
			return fmt.Errorf("Fail to open log file %v", err)
		}
		w = io.MultiWriter(os.Stdout, _logFile)
	} else {
		w = os.Stdout
	}

	log.SetOutput(w)
	_info = log.New(w, "[INFO]: ", log.LUTC|log.Ldate|log.Ltime|log.Lshortfile)
	if debug {
		_debug = log.New(w, "[DEBG]: ", log.LUTC|log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		_debug = nil
	}
	_warning = log.New(w, "[WARN]: ", log.LUTC|log.Ldate|log.Ltime|log.Lshortfile)
	_error = log.New(w, "[EROR]: ", log.LUTC|log.Ldate|log.Ltime|log.Lshortfile)

	return nil
}

// Close is meant to be on programs shutdown to properly close log file used
func Close() {
	_logFile.Close()
}

func Debugf(format string, v ...interface{}) {
	// Make actual write on separate goroutine
	if _debug != nil {
		s := fmt.Sprintf(format, v...)
		_debug.Output(2, s)
	}
}

func Infof(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	_info.Output(2, s)
}

func Warnf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	_warning.Output(2, s)
}

func Errorf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	_error.Output(2, s)
}

func Debug(s string) {
	if _debug != nil {
		_debug.Output(2, s)
	}
}

func Info(s string) {
	_info.Output(2, s)
}

func Warn(s string) {
	_warning.Output(2, s)
}

func Error(s string) {
	_error.Output(2, s)
}
