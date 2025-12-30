package rgrpc

import "sync/atomic"

var defaultCfg atomic.Value // stores Config

// SetDefaultConfig sets the default configuration used by NewClient.
// This is a global setting and should typically be called once at application startup.
// If not called, DefaultConfig() values are used.
func SetDefaultConfig(cfg Config) {
	defaultCfg.Store(cfg)
}

func getDefaultConfig() Config {
	v := defaultCfg.Load()
	if v == nil {
		return DefaultConfig()
	}
	return v.(Config)
}
