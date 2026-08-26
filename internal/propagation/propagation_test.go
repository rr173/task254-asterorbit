package propagation

import (
	"math"
	"testing"
	"time"
)

func TestCircularOrbitKeepsRadius(t *testing.T) {
	el := OrbitElements{A: 1.2, E: 0, EpochJD: 2451545}
	for _, jd := range []float64{2451545, 2451600, 2452000} {
		p := Propagate(el, jd)
		if got := math.Sqrt(p.X*p.X + p.Y*p.Y + p.Z*p.Z); math.Abs(got-el.A) > 1e-10 {
			t.Fatalf("radius at JD %.1f = %.12f, want %.1f", jd, got, el.A)
		}
	}
}

func TestAngularSeparationWrapsRightAscension(t *testing.T) {
	if got := AngularSepArcsec(359.999, 10, 0.001, 10); got > 8 {
		t.Fatalf("RA wrap produced an implausibly large separation: %.3f arcsec", got)
	}
	_, _ = ComputeTopocentric(OrbitElements{A: 1.1, E: .1, EpochJD: 2460000}, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), 35, -120, 1000, 0, 0, 0)
}
