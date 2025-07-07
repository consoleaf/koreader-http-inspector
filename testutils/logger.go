package testutils

import (
	"bytes"
	"log"
	"testing"
)

type testLoggerWriter struct {
	t *testing.T
}

func (w *testLoggerWriter) Write(p []byte) (n int, err error) {
	w.t.Logf("%s", bytes.TrimSuffix(p, []byte("\n")))
	return len(p), nil
}

func MakeTestLogger(t *testing.T) *log.Logger {
	writer := &testLoggerWriter{t: t}
	logger := log.New(writer, "", log.Lshortfile|log.Lmsgprefix)
	return logger
}
