package model

import (
	"errors"
	"testing"
)

func TestObservationValidationBoundaries(t *testing.T) {
	o := Observation{ID: "obs", ArcID: "arc", StationID: "sta", CatalogID: "cat", RA: 0, Dec: -90, Status: ObsStatusRaw}
	if err := o.Validate(); err != nil {
		t.Fatalf("lower sky-coordinate boundary rejected: %v", err)
	}
	o.RA, o.Dec = 359.999999, 90
	if err := o.Validate(); err != nil {
		t.Fatalf("upper sky-coordinate boundary rejected: %v", err)
	}
	o.RA = 360
	if !errors.Is(o.Validate(), ErrBadCoord) {
		t.Fatalf("RA=360 should be rejected as a bad coordinate")
	}
}

func TestArcValidationAndSealedState(t *testing.T) {
	a := ObservationArc{ID: "arc", Name: "test", ObjectName: "2026 NEA", Status: ArcStatusReceiving}
	if err := a.Validate(); err != nil || a.Sealed() {
		t.Fatalf("receiving arc validation/state mismatch: err=%v sealed=%v", err, a.Sealed())
	}
	a.Status = ArcStatusSealed
	if err := a.Validate(); err != nil || !a.Sealed() {
		t.Fatalf("sealed arc validation/state mismatch: err=%v sealed=%v", err, a.Sealed())
	}
}

func TestArcLifecycleTransitions(t *testing.T) {
	valid := [][2]string{
		{ArcStatusReceiving, ArcStatusPendingFit},
		{ArcStatusPendingFit, ArcStatusNeedsReview},
		{ArcStatusNeedsReview, ArcStatusPublished},
		{ArcStatusPublished, ArcStatusSealed},
	}
	for _, pair := range valid {
		if !CanAdvanceArc(pair[0], pair[1]) {
			t.Fatalf("valid transition rejected: %s -> %s", pair[0], pair[1])
		}
	}
	for _, pair := range [][2]string{{ArcStatusReceiving, ArcStatusPublished}, {ArcStatusSealed, ArcStatusReceiving}, {ArcStatusReceiving, "unknown"}} {
		if CanAdvanceArc(pair[0], pair[1]) {
			t.Fatalf("invalid transition accepted: %s -> %s", pair[0], pair[1])
		}
	}
}
