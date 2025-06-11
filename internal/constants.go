package internal

import "time"

const (
	DefaultTTLMinutes    = 10
	MinTTLMinutes        = 1
	MaxTTLMinutes        = 1440

	MaxCodeFailures      = 5
	BlockDuration        = time.Minute

	CleanupInterval      = 5 * time.Minute

	RateLimiterRate      = 5
	RateLimiterBurst     = 10
)
