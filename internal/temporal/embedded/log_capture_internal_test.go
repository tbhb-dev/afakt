// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package embedded

import (
	"log/slog"
	"strings"
	"sync"
	"testing"
)

// logCapture serializes writes and string snapshots of test logs.
type logCapture struct {
	mu  sync.Mutex
	buf []byte
}

// Write appends a log record while excluding readers and other writers.
func (c *logCapture) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.buf = append(c.buf, p...)
	return len(p), nil
}

// String returns a snapshot that remains valid after further writes.
func (c *logCapture) String() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return string(c.buf)
}

func TestLogCaptureConcurrent(t *testing.T) {
	t.Parallel()
	var buf logCapture
	const records = 1000
	start := make(chan struct{})
	var wg sync.WaitGroup
	for range 2 {
		// Separate handlers exercise the capture's lock independently of slog's.
		logger := slog.New(slog.NewTextHandler(&buf, nil))
		wg.Go(func() {
			<-start
			for range records {
				logger.InfoContext(t.Context(), "capture record")
			}
		})
	}
	wg.Go(func() {
		<-start
		for range records {
			if got := buf.String(); got != "" && !strings.HasSuffix(got, "\n") {
				t.Errorf("incomplete log snapshot: %q", got)
			}
		}
	})
	close(start)
	wg.Wait()
	if got := strings.Count(buf.String(), "msg=\"capture record\"\n"); got != 2*records {
		t.Errorf("captured %d records, want %d", got, 2*records)
	}
	snapshot := buf.String()
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	logger.InfoContext(t.Context(), "later record")
	if strings.Contains(snapshot, "later record") {
		t.Error("later write changed an earlier snapshot")
	}
}
