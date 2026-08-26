package store_test

import (
	"errors"
	"path/filepath"
	"testing"

	"task254-asterorbit/internal/model"
	"task254-asterorbit/internal/service"
	"task254-asterorbit/internal/store"
)

func TestArcLifecycleRejectsIllegalJumpsAndUnknownStates(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "asterorbit.db"))
	if err != nil { t.Fatal(err) }
	defer db.Close()
	svc := service.New(db)
	arc, err := svc.CreateArc("arc", "2026 NEA")
	if err != nil { t.Fatal(err) }
	if err := svc.AdvanceArc(arc.ID, model.ArcStatusPublished); !errors.Is(err, model.ErrInvalidArc) { t.Fatalf("receiving to published jump was accepted: %v", err) }
	if err := svc.AdvanceArc(arc.ID, "retired"); !errors.Is(err, model.ErrInvalidArc) { t.Fatalf("unknown state was accepted: %v", err) }
	if err := svc.AdvanceArc(arc.ID, model.ArcStatusPendingFit); err != nil { t.Fatal(err) }
}
