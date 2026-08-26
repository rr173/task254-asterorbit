package service_test

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"task254-asterorbit/internal/model"
	"task254-asterorbit/internal/service"
	"task254-asterorbit/internal/store"
)

func TestSealedArcRejectsNewObservation(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "asterorbit.db"))
	if err != nil { t.Fatal(err) }
	defer db.Close()
	svc := service.New(db)
	cat, _ := svc.CreateCatalog("catalog", "J2000", "ICRS", 0, 0)
	sta, _ := svc.CreateStation("station", "STA", 35, -120, 1000)
	arc, _ := svc.CreateArc("arc", "2026 NEA")
	for _, status := range []string{model.ArcStatusPendingFit, model.ArcStatusNeedsReview, model.ArcStatusPublished} {
		if err := svc.AdvanceArc(arc.ID, status); err != nil { t.Fatal(err) }
	}
	if err := svc.SealArc(arc.ID); err != nil { t.Fatal(err) }
	_, err = svc.ImportObservation(arc.ID, sta.ID, cat.ID, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), 10, 20, .2, .2)
	if !errors.Is(err, model.ErrSealedArc) { t.Fatalf("sealed arc accepted modification: %v", err) }
}
