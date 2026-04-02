package logger

import (
	"io"
	"log"
	"os"
)

type Logger struct {
	verbose bool
	silent  bool
	base    *log.Logger
}

func New(verbose, silent bool, logPath string) (*Logger, error) {
	writers := []io.Writer{os.Stdout}
	if logPath != "" {
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return nil, err
		}
		writers = append(writers, f)
	}

	return &Logger{
		verbose: verbose,
		silent:  silent,
		base:    log.New(io.MultiWriter(writers...), "", log.LstdFlags),
	}, nil
}

func (l *Logger) Info(msg string, args ...any) {
	if l.silent {
		return
	}
	l.base.Printf("level=info msg=%q args=%v", msg, args)
}

func (l *Logger) Verbose(msg string, args ...any) {
	if !l.verbose || l.silent {
		return
	}
	l.base.Printf("level=debug msg=%q args=%v", msg, args)
}

func (l *Logger) Error(msg string, args ...any) {
	l.base.Printf("level=error msg=%q args=%v", msg, args)
}
