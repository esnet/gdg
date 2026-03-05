package config_domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetRetryTimeout_EmptyDelayDefaultsTo100ms(t *testing.T) {
	g := &AppGlobals{RetryDelay: ""}
	d := g.GetRetryTimeout()
	assert.Equal(t, 100*time.Millisecond, d)
}

func TestGetRetryTimeout_ParsesValidDuration(t *testing.T) {
	g := &AppGlobals{RetryDelay: "500ms"}
	d := g.GetRetryTimeout()
	assert.Equal(t, 500*time.Millisecond, d)
}

func TestGetRetryTimeout_ParsesSeconds(t *testing.T) {
	g := &AppGlobals{RetryDelay: "2s"}
	d := g.GetRetryTimeout()
	assert.Equal(t, 2*time.Second, d)
}

func TestGetRetryTimeout_InvalidStringDefaultsTo100ms(t *testing.T) {
	g := &AppGlobals{RetryDelay: "not-a-duration"}
	d := g.GetRetryTimeout()
	assert.Equal(t, 100*time.Millisecond, d)
}

func TestGetRetryTimeout_CachesResult(t *testing.T) {
	g := &AppGlobals{RetryDelay: "200ms"}
	d1 := g.GetRetryTimeout()
	// Change the string — cached value should still be returned
	g.RetryDelay = "999s"
	d2 := g.GetRetryTimeout()
	assert.Equal(t, d1, d2, "second call should return the cached value")
}
