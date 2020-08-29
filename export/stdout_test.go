package export

import (
	"bytes"
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
)

var (
	stdoutExporter = &Stdout{
		config: config,
	}
)

func TestNewStdout(t *testing.T) {
	got, err := NewStdout(config)
	got.Stop()
	if err != nil {
		t.Errorf("Error creating new Stdout")
	}

	gotStdout := got.Exporter.(Stdout)
	compareStdout(t, *stdoutExporter, gotStdout)
}

func TestStdoutExportMetrics(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)

	stdoutExporter.ExportMetrics(context.Background(), metrics)

	log.SetOutput(os.Stderr)
	if !strings.Contains(buf.String(), dummyName) {
		t.Fatalf("Stdout Export Metrics failed, could not find %v", dummyName)
	}

	if !strings.Contains(buf.String(), dummyDesc) {
		t.Fatalf("Stdout Export Metrics failed, could not find %v", dummyDesc)
	}

	if !strings.Contains(buf.String(), strconv.Itoa(int(intVal))) {
		t.Fatalf("Stdout Export Metrics failed, could not find %v", strconv.Itoa(int(intVal)))
	}
}

func compareStdout(t *testing.T, want Stdout, got Stdout) {
	if want.config != got.config {
		t.Fatalf("New Stdout failed, expected config %v, got %v", want.config, got.config)
	}
}
