package domain

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/lmittmann/tint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tmpLogFile(t *testing.T) *os.File {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "log-*.txt")
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	return f
}

func readLogFile(t *testing.T, f *os.File) string {
	t.Helper()
	_, err := f.Seek(0, io.SeekStart)
	require.NoError(t, err)
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, f)
	return buf.String()
}

func newLogHandler(t *testing.T) (*ContextHandler, *os.File, *os.File) {
	t.Helper()
	outFile := tmpLogFile(t)
	errFile := tmpLogFile(t)
	opts := &tint.Options{Level: slog.LevelDebug}
	h := NewContextHandler(nil, outFile, errFile, opts)
	require.NotNil(t, h)
	return h, outFile, errFile
}

// ── NewContextHandler ─────────────────────────────────────────────────────────

func TestNewContextHandler_ReturnsNonNil(t *testing.T) {
	h, _, _ := newLogHandler(t)
	assert.NotNil(t, h)
}

func TestNewContextHandler_ReturnsSameHandlerWhenParamsMatch(t *testing.T) {
	outFile := tmpLogFile(t)
	errFile := tmpLogFile(t)
	opts := &tint.Options{Level: slog.LevelDebug}

	h1 := NewContextHandler(nil, outFile, errFile, opts)
	// Passing h1 back with the same out/err/level → should return h1 (identity check).
	h2 := NewContextHandler(h1, outFile, errFile, opts)
	assert.Equal(t, h1, h2, "should return the cached ContextHandler when params match")
}

func TestNewContextHandler_CreatesNewHandlerWhenLevelDiffers(t *testing.T) {
	outFile := tmpLogFile(t)
	errFile := tmpLogFile(t)
	opts1 := &tint.Options{Level: slog.LevelDebug}
	opts2 := &tint.Options{Level: slog.LevelInfo}

	h1 := NewContextHandler(nil, outFile, errFile, opts1)
	h2 := NewContextHandler(h1, outFile, errFile, opts2)
	assert.NotEqual(t, h1, h2, "different level should produce a new handler")
}

// ── Enabled ───────────────────────────────────────────────────────────────────

func TestEnabled_DebugIsEnabledWhenHandlerAtDebugLevel(t *testing.T) {
	h, _, _ := newLogHandler(t)
	assert.True(t, h.Enabled(context.Background(), slog.LevelDebug))
}

func TestEnabled_WarnIsEnabled(t *testing.T) {
	h, _, _ := newLogHandler(t)
	assert.True(t, h.Enabled(context.Background(), slog.LevelWarn))
}

func TestEnabled_ErrorIsEnabled(t *testing.T) {
	h, _, _ := newLogHandler(t)
	assert.True(t, h.Enabled(context.Background(), slog.LevelError))
}

// ── Handle — routing ──────────────────────────────────────────────────────────

func TestHandle_InfoRecordGoesToOutStream(t *testing.T) {
	h, outFile, errFile := newLogHandler(t)
	ctx := context.Background()

	rec := slog.NewRecord(time.Now(), slog.LevelInfo, "hello info", 0)
	err := h.Handle(ctx, rec)
	assert.NoError(t, err)

	assert.Contains(t, readLogFile(t, outFile), "hello info")
	assert.NotContains(t, readLogFile(t, errFile), "hello info")
}

func TestHandle_WarnRecordGoesToErrStream(t *testing.T) {
	h, outFile, errFile := newLogHandler(t)
	ctx := context.Background()

	rec := slog.NewRecord(time.Now(), slog.LevelWarn, "warn message", 0)
	err := h.Handle(ctx, rec)
	assert.NoError(t, err)

	assert.NotContains(t, readLogFile(t, outFile), "warn message")
	assert.Contains(t, readLogFile(t, errFile), "warn message")
}

func TestHandle_ErrorRecordGoesToErrStream(t *testing.T) {
	h, outFile, errFile := newLogHandler(t)
	ctx := context.Background()

	rec := slog.NewRecord(time.Now(), slog.LevelError, "error occurred", 0)
	err := h.Handle(ctx, rec)
	assert.NoError(t, err)

	assert.NotContains(t, readLogFile(t, outFile), "error occurred")
	assert.Contains(t, readLogFile(t, errFile), "error occurred")
}

func TestHandle_DebugRecordGoesToOutStream(t *testing.T) {
	h, outFile, errFile := newLogHandler(t)
	ctx := context.Background()

	rec := slog.NewRecord(time.Now(), slog.LevelDebug, "debug detail", 0)
	err := h.Handle(ctx, rec)
	assert.NoError(t, err)

	assert.Contains(t, readLogFile(t, outFile), "debug detail")
	assert.NotContains(t, readLogFile(t, errFile), "debug detail")
}
