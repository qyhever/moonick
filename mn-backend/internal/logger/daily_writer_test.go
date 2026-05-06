package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDailyLogWriterCreatesDatedFile(t *testing.T) {
	now := time.Date(2026, 5, 6, 10, 30, 0, 0, time.Local)
	writer := newDailyLogWriter(filepath.Join(t.TempDir(), "moonick.log"), 0, 0)
	writer.now = func() time.Time { return now }

	if _, err := writer.Write([]byte("hello")); err != nil {
		t.Fatalf("write log: %v", err)
	}

	filename := filepath.Join(filepath.Dir(writer.baseFilename), "moonick-2026-05-06.log")
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read dated log file: %v", err)
	}
	if string(content) != "hello" {
		t.Fatalf("dated log content = %q, want hello", string(content))
	}
}

func TestDailyLogWriterRotatesWhenDateChanges(t *testing.T) {
	current := time.Date(2026, 5, 6, 23, 59, 0, 0, time.Local)
	writer := newDailyLogWriter(filepath.Join(t.TempDir(), "moonick.log"), 0, 0)
	writer.now = func() time.Time { return current }

	if _, err := writer.Write([]byte("day-one")); err != nil {
		t.Fatalf("write first day log: %v", err)
	}

	current = current.Add(2 * time.Minute)

	if _, err := writer.Write([]byte("day-two")); err != nil {
		t.Fatalf("write second day log: %v", err)
	}

	firstFile := filepath.Join(filepath.Dir(writer.baseFilename), "moonick-2026-05-06.log")
	secondFile := filepath.Join(filepath.Dir(writer.baseFilename), "moonick-2026-05-07.log")

	firstContent, err := os.ReadFile(firstFile)
	if err != nil {
		t.Fatalf("read first log file: %v", err)
	}
	if string(firstContent) != "day-one" {
		t.Fatalf("first dated log content = %q, want day-one", string(firstContent))
	}

	secondContent, err := os.ReadFile(secondFile)
	if err != nil {
		t.Fatalf("read second log file: %v", err)
	}
	if string(secondContent) != "day-two" {
		t.Fatalf("second dated log content = %q, want day-two", string(secondContent))
	}
}

func TestDailyLogWriterRemovesOldBackups(t *testing.T) {
	dir := t.TempDir()
	writeLogFixture(t, filepath.Join(dir, "moonick-2026-05-04.log"))
	writeLogFixture(t, filepath.Join(dir, "moonick-2026-05-05.log"))

	now := time.Date(2026, 5, 6, 10, 30, 0, 0, time.Local)
	writer := newDailyLogWriter(filepath.Join(dir, "moonick.log"), 2, 0)
	writer.now = func() time.Time { return now }

	if _, err := writer.Write([]byte("latest")); err != nil {
		t.Fatalf("write latest log: %v", err)
	}

	files, err := filepath.Glob(filepath.Join(dir, "moonick-*.log"))
	if err != nil {
		t.Fatalf("glob dated log files: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("dated log file count = %d, want 2", len(files))
	}
	if strings.Contains(strings.Join(files, ","), "2026-05-04") {
		t.Fatalf("oldest dated log file should be removed, files = %v", files)
	}
}

func writeLogFixture(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(path), 0o644); err != nil {
		t.Fatalf("write fixture %s: %v", path, err)
	}
}
