package service_test

import (
	"path/filepath"
	"testing"
	"time"

	"task254-asterorbit/internal/service"
	"task254-asterorbit/internal/store"
)

func TestSameInstantInDifferentTimeZonesIsIdempotent(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "asterorbit.db"))
	if err != nil { t.Fatal(err) }
	defer db.Close()
	svc := service.New(db)
	cat, _ := svc.CreateCatalog("catalog", "J2000", "ICRS", 0, 0)
	sta, _ := svc.CreateStation("station", "STA", 35, -120, 1000)
	arc, _ := svc.CreateArc("arc", "2026 NEA")
	utc := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	local := utc.In(time.FixedZone("UTC+8", 8*60*60))
	first, err := svc.ImportObservation(arc.ID, sta.ID, cat.ID, utc, 10, 20, .2, .2)
	if err != nil { t.Fatal(err) }
	second, err := svc.ImportObservation(arc.ID, sta.ID, cat.ID, local, 10, 20, .2, .2)
	if err != nil { t.Fatal(err) }
	if first.ID != second.ID { t.Fatalf("same instant received different IDs: %s vs %s", first.ID, second.ID) }
	items, err := svc.ListObservations(arc.ID)
	if err != nil { t.Fatal(err) }
	if len(items) != 1 { t.Fatalf("same instant was stored more than once: %d rows", len(items)) }
}
