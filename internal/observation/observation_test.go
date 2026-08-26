package observation

import (
	"testing"
	"time"

	"task254-asterorbit/internal/model"
)

func TestDeriveObservationIDIsStableAndInputSensitive(t *testing.T) {
	tm := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	a := DeriveObservationID("arc", "sta", "cat", tm, 12.3, -4.5)
	if a != DeriveObservationID("arc", "sta", "cat", tm, 12.3, -4.5) {
		t.Fatal("same observation did not derive the same ID")
	}
	if a == DeriveObservationID("arc", "sta", "cat", tm, 12.4, -4.5) {
		t.Fatal("different coordinates derived the same ID")
	}
}

// TestDeriveObservationIDIsTimeZoneInvariant 锁定幂等行为：同一物理时刻无论以哪个
// 时区文字表示都应派生同一 ID，避免重复导入新建重复测角记录；而真正不同的时刻
// 不得互相覆盖。
func TestDeriveObservationIDIsTimeZoneInvariant(t *testing.T) {
	tUTC := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	// 同一时刻的几种等价时区表示
	tJST := tUTC.In(time.FixedZone("JST", 9*3600))
	tEST := tUTC.In(time.FixedZone("EST", -5*3600))
	tPST := tUTC.In(time.FixedZone("PST", -8*3600))

	want := DeriveObservationID("arc", "sta", "cat", tUTC, 12.3, -4.5)
	for _, eq := range []time.Time{tJST, tEST, tPST} {
		if got := DeriveObservationID("arc", "sta", "cat", eq, 12.3, -4.5); got != want {
			t.Fatalf("same instant via zone %s derived %s, want %s (idempotency broken)",
				eq.Location().String(), got, want)
		}
	}

	// 真正不同的时刻（相差 1 秒）必须派生不同 ID，不得互相覆盖。
	other := DeriveObservationID("arc", "sta", "cat", tUTC.Add(1*time.Second), 12.3, -4.5)
	if other == want {
		t.Fatal("observations one second apart derived the same ID (distinct observations would collapse)")
	}

	// 带亚秒纳秒的同一时刻，不同表示仍须等价。
	withNanos := time.Date(2026, 3, 1, 0, 0, 0, 500000000, time.UTC)
	wantNano := DeriveObservationID("arc", "sta", "cat", withNanos, 1.0, 1.0)
	if got := DeriveObservationID("arc", "sta", "cat", withNanos.In(time.FixedZone("CET", 3600)), 1.0, 1.0); got != wantNano {
		t.Fatalf("sub-second instant via different zone derived %s, want %s", got, wantNano)
	}
}

func TestFlagAnomalousUsesStrictThreshold(t *testing.T) {
	if FlagAnomalous(30, 30) {
		t.Fatal("a residual exactly at the threshold should not be anomalous")
	}
	if !FlagAnomalous(30.001, 30) {
		t.Fatal("a residual above the threshold should be anomalous")
	}
	if err := ValidateObservation(model.Observation{ID: "o", ArcID: "a", StationID: "s", CatalogID: "c", RA: 1, Dec: 2, Status: model.ObsStatusRaw}); err != nil {
		t.Fatalf("valid observation rejected: %v", err)
	}
}
