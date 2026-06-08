package timeutil

import (
	"testing"
	"time"
)

func TestNow_ReturnsUTC(t *testing.T) {
	got := Now()
	if got.Location() != time.UTC {
		t.Errorf("Now() returned time with location %v, want UTC", got.Location())
	}
}

func TestNow_NotMoreThanMaxTimeDriftInFuture(t *testing.T) {
	now := time.Now().UTC()
	maxAllowed := now.Add(MaxTimeDrift)

	got := Now()

	if got.After(maxAllowed) {
		t.Errorf("Now() = %v, is more than MaxTimeDrift (%v) in the future from reference now (%v)", got, MaxTimeDrift, now)
	}
}

func TestNow_ReturnsTimeWithinExpectedBounds(t *testing.T) {
	// Capture a reference time before calling Now()
	before := time.Now().UTC()

	got := Now()

	// The result should not be significantly in the past (should be close to "now")
	// and should be within MaxTimeDrift of the reference
	minExpected := before.Add(-time.Second) // Allow 1 second tolerance for test execution
	maxExpected := before.Add(MaxTimeDrift)

	if got.Before(minExpected) {
		t.Errorf("Now() = %v, is unexpectedly far in the past (before %v)", got, minExpected)
	}
	if got.After(maxExpected) {
		t.Errorf("Now() = %v, is more than MaxTimeDrift (%v) in the future (after %v)", got, MaxTimeDrift, maxExpected)
	}
}

func TestNow_CalledMultipleTimes_ReturnsConsistentResults(t *testing.T) {
	times := make([]time.Time, 5)

	for i := 0; i < 5; i++ {
		times[i] = Now()
	}

	// All times should be in UTC
	for i, tm := range times {
		if tm.Location() != time.UTC {
			t.Errorf("Now() call %d returned time with location %v, want UTC", i+1, tm.Location())
		}
	}

	// All times should be within MaxTimeDrift of each other
	for i := 1; i < len(times); i++ {
		diff := times[i].Sub(times[0])
		if diff < 0 {
			diff = -diff
		}
		if diff > MaxTimeDrift {
			t.Errorf("Now() call %d and call 0 differ by %v, want difference <= MaxTimeDrift (%v)", i, diff, MaxTimeDrift)
		}
	}
}

func TestMaxTimeDrift_Constant(t *testing.T) {
	// Verify MaxTimeDrift is defined as expected
	if MaxTimeDrift != time.Minute {
		t.Errorf("MaxTimeDrift = %v, want %v", MaxTimeDrift, time.Minute)
	}
}