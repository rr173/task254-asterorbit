package observation_test

import (
	"math"
	"testing"
	"time"

	"task254-asterorbit/internal/model"
	"task254-asterorbit/internal/observation"
	"task254-asterorbit/internal/propagation"
)

func TestResidualWrapsAcrossZeroRightAscension(t *testing.T) {
	el := propagation.OrbitElements{A: 1.05, E: 0.2, IDeg: 12, OmDeg: 75, WDeg: 40, M0Deg: 5, EpochJD: 2461000.5}
	var tm time.Time
	var computedRA, computedDec float64
	for day := 0; day < 2000; day++ {
		candidate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, day)
		computedRA, computedDec = propagation.ComputeTopocentric(el, candidate, 35, -120, 1000, 0, 0, 0)
		if computedRA > 359.5 {
			tm = candidate
			break
		}
	}
	if tm.IsZero() {
		t.Fatal("fixture did not find a deterministic RA wrap case")
	}
	// Deliberately cross 360 degrees by a visible 0.6 degree offset.  The
	// residual should stay at 0.6 degrees, not become almost one full turn.
	observedRA := computedRA + 0.6
	if observedRA >= 360 {
		observedRA -= 360
	}
	residual := observation.ComputeResidual(model.Observation{
		ID: "obs", ArcID: "arc", StationID: "sta", CatalogID: "cat", ObsTimeUTC: tm,
		RA: observedRA, Dec: computedDec, Status: model.ObsStatusRaw,
	}, model.Orbit{ID: "orb", ArcID: "arc", A: el.A, E: el.E, IDeg: el.IDeg, OmDeg: el.OmDeg, WDeg: el.WDeg, M0Deg: el.M0Deg, EpochJD: el.EpochJD}, model.Station{
		ID: "sta", Name: "station", Code: "STA", LatitudeDeg: 35, LongitudeDeg: -120, HeightM: 1000,
	})
	if math.Abs(math.Abs(residual.ResRA*3600)-2160) > 1.0 {
		t.Fatalf("RA wrap residual is too large: computed=%.9f observed=%.9f res=%.6f arcsec", computedRA, observedRA, residual.ResRA*3600)
	}
}
