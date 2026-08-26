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

func TestAnomalousObservationsRemainInRecompute(t *testing.T) {
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
	tm := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	el := propagation.OrbitElements{A: orb.A, E: orb.E, IDeg: orb.IDeg, OmDeg: orb.OmDeg, WDeg: orb.WDeg, M0Deg: orb.M0Deg, EpochJD: orb.EpochJD}
	ra, dec := propagation.ComputeTopocentric(el, tm, 35, -120, 1000, 0, 0, 0)
	if _, err := svc.ImportObservation(arc.ID, sta.ID, cat.ID, tm, ra+40.0/3600, dec, .2, .2); err != nil {
		t.Fatal(err)
	}
	if n, err := svc.ComputeResiduals(arc.ID); err != nil || n != 1 {
		t.Fatalf("first compute = %d, %v", n, err)
	}
	obs, err := svc.ListObservations(arc.ID)
	if err != nil || len(obs) != 1 || obs[0].Status != model.ObsStatusAnomalous {
		t.Fatalf("observation was not marked anomalous: len=%d err=%v status=%v", len(obs), err, obs[0].Status)
	}
	if n, err := svc.ComputeResiduals(arc.ID); err != nil || n != 1 {
		t.Fatalf("anomalous observation disappeared on recompute: n=%d err=%v", n, err)
	}
}
