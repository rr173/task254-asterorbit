package propagation_test

import (
	"math"
	"testing"
	"time"

	"task254-asterorbit/internal/propagation"
)

func TestStationLongitudeFixture(t *testing.T) {
	el := propagation.OrbitElements{A: 1.1, E: 0.2, IDeg: 12, OmDeg: 75, WDeg: 40, M0Deg: 5, EpochJD: 2461000.5}
	ra, dec := propagation.ComputeTopocentric(el, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), 35, 120, 1000, 0, 0, 0)
	if math.Abs(ra-279.014180461830563) > 1e-9 || math.Abs(dec+17.967530639640678) > 1e-9 {
		t.Fatalf("station longitude conversion changed the predicted direction: ra=%.15f dec=%.15f", ra, dec)
	}
}
