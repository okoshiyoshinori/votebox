package logger

import (
	"log"
	"os"
)

var (
	Debug   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func Newlogger(f *os.File, fe *os.File) {
	Info = log.New(f, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(f, "[WARN] ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(fe, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
}
