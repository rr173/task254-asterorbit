package service

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"task254-asterorbit/internal/model"
	"task254-asterorbit/internal/store"
)

// newTestDB 打开一个临时 SQLite 数据库，供 service 层集成测试使用。
func newTestDB(t *testing.T) *store.DB {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// seedArc 固件：创建一个台站、一个星表与一条已封存的弧段，并返回它们。
func seedSealedArc(t *testing.T, s *Service) (model.ObservationArc, model.Station, model.StarCatalog) {
	t.Helper()
	st, err := s.CreateStation("Mount Test", "TST", 35.0, -120.0, 1700)
	if err != nil {
		t.Fatalf("create station: %v", err)
	}
	cat, err := s.CreateCatalog("Test Catalog", "J2000", "ICRS", 0, 0)
	if err != nil {
		t.Fatalf("create catalog: %v", err)
	}
	arc, err := s.CreateArc("sealed arc", "2026 TEST")
	if err != nil {
		t.Fatalf("create arc: %v", err)
	}
	// 推进到 published 后封存。
	if err := s.DB.UpdateArcStatus(arc.ID, model.ArcStatusPendingFit); err != nil {
		t.Fatalf("advance to pending_fit: %v", err)
	}
	if err := s.DB.UpdateArcStatus(arc.ID, model.ArcStatusNeedsReview); err != nil {
		t.Fatalf("advance to needs_review: %v", err)
	}
	if err := s.DB.UpdateArcStatus(arc.ID, model.ArcStatusPublished); err != nil {
		t.Fatalf("advance to published: %v", err)
	}
	if err := s.SealArc(arc.ID); err != nil {
		t.Fatalf("seal arc: %v", err)
	}
	return arc, st, cat
}

// TestImportObservationRejectsSealedArc 验证封存弧段禁止再导入测角记录，
// 防止封存快照的审计边界被破坏。
func TestImportObservationRejectsSealedArc(t *testing.T) {
	svc := New(newTestDB(t))
	arc, st, cat := seedSealedArc(t, svc)

	tm := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	_, err := svc.ImportObservation(arc.ID, st.ID, cat.ID, tm, 12.3, -4.5, 0.3, 0.3)
	if !errors.Is(err, model.ErrSealedArc) {
		t.Fatalf("importing into a sealed arc must be refused with ErrSealedArc, got %v", err)
	}

	// 封存弧段的观测数不应增长。
	n, err := svc.DB.CountObservationsByArc(arc.ID)
	if err != nil {
		t.Fatalf("count observations: %v", err)
	}
	if n != 0 {
		t.Fatalf("sealed arc should have no observations, got %d", n)
	}
}
