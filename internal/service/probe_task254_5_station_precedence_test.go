package service_test

import (
	"path/filepath"
	"testing"
	"time"

	"task254-asterorbit/internal/model"
	"task254-asterorbit/internal/propagation"
	"task254-asterorbit/internal/service"
	"task254-asterorbit/internal/store"
)

func TestEqualStationAndCatalogSignalsPreferStation(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "asterorbit.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	svc := service.New(db)
	cat, _ := svc.CreateCatalog("catalog", "J2000", "ICRS", 0, 0)
	sta, _ := svc.CreateStation("station", "STA", 35, -120, 1000)
	arc, _ := svc.CreateArc("arc", "2026 NEA")
	orb, err := svc.SetOrbit(arc.ID, 1.1, .15, 12, 75, 40, 5, 2461000.5, "probe")
	if err != nil {
		t.Fatal(err)
	}
	el := propagation.OrbitElements{A: orb.A, E: orb.E, IDeg: orb.IDeg, OmDeg: orb.OmDeg, WDeg: orb.WDeg, M0Deg: orb.M0Deg, EpochJD: orb.EpochJD}
	for i := 0; i < 3; i++ {
		tm := time.Date(2026, 3, 1+i, 0, 0, 0, 0, time.UTC)
		ra, dec := propagation.ComputeTopocentric(el, tm, 35, -120, 1000, 0, 0, 0)
		if _, err := svc.ImportObservation(arc.ID, sta.ID, cat.ID, tm, ra+40.0/3600, dec, .2, .2); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := svc.ComputeResiduals(arc.ID); err != nil {
		t.Fatal(err)
	}
	a, err := svc.Analyze(arc.ID)
	if err != nil {
		t.Fatal(err)
	}
	if a.Kind != model.AttrKindStation || a.TargetID != sta.ID {
		t.Fatalf("equal signals were not attributed to station: kind=%s target=%s", a.Kind, a.TargetID)
	}
}
