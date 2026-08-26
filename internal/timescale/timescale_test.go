package timescale

import (
	"testing"
	"time"
)

func TestTAIOffsetAcrossLeapSecond(t *testing.T) {
	before := time.Date(2016, 12, 31, 23, 59, 59, 0, time.UTC)
	after := time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)
	if got := TAIOffsetSeconds(before); got != 36 {
		t.Fatalf("offset before 2017 leap-second boundary = %d, want 36", got)
	}
	if got := TAIOffsetSeconds(after); got != 37 {
		t.Fatalf("offset at 2017 leap-second boundary = %d, want 37", got)
	}
	if UTCToTAI(before).Sub(before) != 36*time.Second {
		t.Fatalf("UTC to TAI conversion did not apply the historical offset")
	}
}

func TestJDRoundTripUsesNanoseconds(t *testing.T) {
	tm := time.Date(2026, 3, 1, 12, 34, 56, 789000000, time.UTC)
	jd := JDFromTime(tm)
	if delta := (jd - JDFromTime(tm.Add(500*time.Millisecond))) * 86400; delta >= -0.49 || delta <= -0.51 {
		t.Fatalf("unexpected JD nanosecond resolution: delta seconds=%v", delta)
	}
}
