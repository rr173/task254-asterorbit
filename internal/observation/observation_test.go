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
