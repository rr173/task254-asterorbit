package service_test

import (
	"math"
	"path/filepath"
	"testing"
	"time"

	"task254-asterorbit/internal/propagation"
	"task254-asterorbit/internal/service"
	"task254-asterorbit/internal/store"
)

func TestStationCalibrationReducesResidual(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "asterorbit.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	svc := service.New(db)
	cat, err := svc.CreateCatalog("catalog", "J2000", "ICRS", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	sta, err := svc.CreateStation("station", "STA", 35, -120, 1000)
	if err != nil {
		t.Fatal(err)
	}
	arc, err := svc.CreateArc("arc", "2026 NEA")
	if err != nil {
		t.Fatal(err)
	}
	orb, err := svc.SetOrbit(arc.ID, 1.1, 0.15, 12, 75, 40, 5, 2461000.5, "probe")
	if err != nil {
		t.Fatal(err)
	}
	el := propagation.OrbitElements{A: orb.A, E: orb.E, IDeg: orb.IDeg, OmDeg: orb.OmDeg, WDeg: orb.WDeg, M0Deg: orb.M0Deg, EpochJD: orb.EpochJD}
	const raBias = 36.0
	for i := 0; i < 4; i++ {
		tm := time.Date(2026, 3, 1+i, 0, 0, 0, 0, time.UTC)
		ra, dec := propagation.ComputeTopocentric(el, tm, sta.LatitudeDeg, sta.LongitudeDeg, sta.HeightM, 0, 0, 0)
		if _, err := svc.ImportObservation(arc.ID, sta.ID, cat.ID, tm, ra+raBias/3600, dec, 0.2, 0.2); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := svc.ComputeResiduals(arc.ID); err != nil {
		t.Fatal(err)
	}
	before := rms(t, svc, arc.ID)
	if err := svc.ApplyCalibration(sta.ID, raBias, 0, "probe"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ComputeResiduals(arc.ID); err != nil {
		t.Fatal(err)
	}
	after := rms(t, svc, arc.ID)
	if !(after < before*0.1) {
		t.Fatalf("calibration did not converge residuals: before=%.6f after=%.6f", before, after)
	}
}

func rms(t *testing.T, svc *service.Service, arcID string) float64 {
	t.Helper()
	items, err := svc.GetResiduals(arcID)
	if err != nil {
		t.Fatal(err)
	}
	var sum float64
	for _, item := range items {
		sum += item.SepArcsec * item.SepArcsec
	}
	return math.Sqrt(sum / float64(len(items)))
}
