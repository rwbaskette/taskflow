// Package timeutil provides time utilities to prevent clock drift issues.
package timeutil

import "time"

// MaxTimeDrift is the maximum allowed clock drift into the future.
// If the system clock is set more than 1 minute ahead, we clamp it to now.
const MaxTimeDrift = time.Minute

// Now returns the current UTC time, clamped to prevent future timestamps
// from being set due to system clock drift.
func Now() time.Time {
	now := time.Now().UTC()
	maxAllowed := now.Add(MaxTimeDrift)
	if now.After(maxAllowed) {
		return maxAllowed
	}
	return now
}