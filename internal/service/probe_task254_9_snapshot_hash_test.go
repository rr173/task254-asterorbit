package service_test

import (
	"path/filepath"
	"testing"
	"time"

	"task254-asterorbit/internal/service"
	"task254-asterorbit/internal/store"
)

func TestSameSnapshotContentHasStableHash(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "asterorbit.db"))
	if err != nil { t.Fatal(err) }
	defer db.Close()
	svc := service.New(db)
	cat, _ := svc.CreateCatalog("catalog", "J2000", "ICRS", 0, 0)
	arc, _ := svc.CreateArc("arc", "2026 NEA")
	if _, err := svc.SetOrbit(arc.ID, 1.1, .15, 12, 75, 40, 5, 2461000.5, "probe"); err != nil { t.Fatal(err) }
	first, err := svc.PublishOrbitSnapshot(arc.ID, cat.ID)
	if err != nil { t.Fatal(err) }
	time.Sleep(1100 * time.Millisecond)
	second, err := svc.PublishOrbitSnapshot(arc.ID, cat.ID)
	if err != nil { t.Fatal(err) }
	if first.Hash != second.Hash { t.Fatalf("same scientific content changed hash: first=%s second=%s", first.Hash, second.Hash) }
}
