package logger

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestNewWithWriterEmitsServiceAttr(t *testing.T) {
	var buf bytes.Buffer
	log := NewWithWriter(&buf, "iam")
	log.Info("hello")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("log line is not valid JSON: %v (%q)", err, buf.String())
	}
	if entry["service"] != "iam" {
		t.Errorf("service attr = %v; want iam", entry["service"])
	}
	if entry["msg"] != "hello" {
		t.Errorf("msg = %v; want hello", entry["msg"])
	}
}

func TestLevelFromEnv(t *testing.T) {
	t.Setenv("LOG_LEVEL", "warn")
	var buf bytes.Buffer
	log := NewWithWriter(&buf, "iam")
	log.Info("should be filtered")
	if buf.Len() != 0 {
		t.Errorf("info log emitted at warn level: %q", buf.String())
	}
	log.Warn("should appear")
	if buf.Len() == 0 {
		t.Error("warn log was filtered at warn level")
	}
}
