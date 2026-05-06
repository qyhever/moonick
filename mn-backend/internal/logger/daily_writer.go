package logger

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const logDateLayout = "2006-01-02"

type dailyLogWriter struct {
	baseFilename string
	maxBackups   int
	maxAge       int
	now          func() time.Time

	mu          sync.Mutex
	file        *os.File
	currentDate string
}

func newDailyLogWriter(filename string, maxBackups, maxAge int) *dailyLogWriter {
	return &dailyLogWriter{
		baseFilename: filename,
		maxBackups:   maxBackups,
		maxAge:       maxAge,
		now:          time.Now,
	}
}

func (w *dailyLogWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.rotateIfNeeded(w.now()); err != nil {
		return 0, err
	}

	return w.file.Write(p)
}

func (w *dailyLogWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}

	return w.file.Sync()
}

func (w *dailyLogWriter) rotateIfNeeded(now time.Time) error {
	date := now.Format(logDateLayout)
	if w.file != nil && w.currentDate == date {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(w.baseFilename), 0o755); err != nil {
		return err
	}

	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return err
		}
	}

	file, err := os.OpenFile(datedFilename(w.baseFilename, now), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	w.file = file
	w.currentDate = date
	w.cleanupOldFiles(now)
	return nil
}

func (w *dailyLogWriter) cleanupOldFiles(now time.Time) {
	entries := listDatedLogFiles(w.baseFilename)
	if len(entries) == 0 {
		return
	}

	if w.maxAge > 0 {
		cutoff := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(w.maxAge - 1))
		filtered := entries[:0]
		for _, entry := range entries {
			if entry.date.Before(cutoff) {
				_ = os.Remove(entry.path)
				continue
			}
			filtered = append(filtered, entry)
		}
		entries = filtered
	}

	if w.maxBackups > 0 && len(entries) > w.maxBackups {
		for _, entry := range entries[:len(entries)-w.maxBackups] {
			_ = os.Remove(entry.path)
		}
	}
}

type datedLogFile struct {
	path string
	date time.Time
}

func listDatedLogFiles(baseFilename string) []datedLogFile {
	pattern := datedFilenameGlob(baseFilename)
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil
	}

	entries := make([]datedLogFile, 0, len(files))
	for _, file := range files {
		date, ok := parseDateFromFilename(baseFilename, file)
		if !ok {
			continue
		}
		entries = append(entries, datedLogFile{path: file, date: date})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].date.Before(entries[j].date)
	})

	return entries
}

func datedFilename(baseFilename string, now time.Time) string {
	dir := filepath.Dir(baseFilename)
	ext := filepath.Ext(baseFilename)
	name := strings.TrimSuffix(filepath.Base(baseFilename), ext)
	return filepath.Join(dir, name+"-"+now.Format(logDateLayout)+ext)
}

func datedFilenameGlob(baseFilename string) string {
	dir := filepath.Dir(baseFilename)
	ext := filepath.Ext(baseFilename)
	name := strings.TrimSuffix(filepath.Base(baseFilename), ext)
	return filepath.Join(dir, name+"-*"+ext)
}

func parseDateFromFilename(baseFilename, filename string) (time.Time, bool) {
	ext := filepath.Ext(baseFilename)
	name := strings.TrimSuffix(filepath.Base(baseFilename), ext)
	base := filepath.Base(filename)
	prefix := name + "-"

	if !strings.HasPrefix(base, prefix) || !strings.HasSuffix(base, ext) {
		return time.Time{}, false
	}

	datePart := strings.TrimSuffix(strings.TrimPrefix(base, prefix), ext)
	date, err := time.ParseInLocation(logDateLayout, datePart, time.Local)
	if err != nil {
		return time.Time{}, false
	}

	return date, true
}
