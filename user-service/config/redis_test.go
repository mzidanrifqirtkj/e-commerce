package config

import (
	"testing"

	"github.com/spf13/viper"
)

// The address must come from REDIS_HOST/REDIS_PORT, not from a literal in the
// source. redis.Client exposes the resolved options, so this reads back the
// exact address the client would dial.
//
// The missing-key branch is deliberately untested here: it calls log.Fatal,
// which would take the whole test binary down with it.
func TestNewRedisClientReadsAddressFromViper(t *testing.T) {
	viper.Set("REDIS_HOST", "cache.internal")
	viper.Set("REDIS_PORT", "6380")

	addr := NewRedisClient().Options().Addr
	if addr != "cache.internal:6380" {
		t.Fatalf("Addr = %q, want %q", addr, "cache.internal:6380")
	}
}