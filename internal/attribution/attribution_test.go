package attribution

import (
	"testing"

	"task254-asterorbit/internal/model"
)

func TestAnalyzeDetectsStationPattern(t *testing.T) {
	items := make([]ObsResidual, 0, 6)
	for i := 0; i < 3; i++ {
		items = append(items, ObsResidual{Residual: model.Residual{ResRA: 0.02, ResDec: 0.001, SepArcsec: 72}, StationID: "biased", CatalogID: "cat", Days: float64(i)})
		items = append(items, ObsResidual{Residual: model.Residual{ResRA: 0, ResDec: 0, SepArcsec: 0}, StationID: "nominal", CatalogID: "cat", Days: float64(i)})
	}
	a := Analyze("arc", items)
	if a.Kind != model.AttrKindStation || a.TargetID != "biased" {
		t.Fatalf("station pattern classified as kind=%s target=%s", a.Kind, a.TargetID)
	}
}

func TestAnalyzePrefersStrongerStationOverCatalog(t *testing.T) {
	// 台站 A 带固定偏移（组内高度一致），星表整体也有偏移（被未偏移台站 B 稀释，
	// 组内一致性弱）。台站信号信噪比更高，应优先归因台站，而非星表。
	items := make([]ObsResidual, 0, 8)
	for i := 0; i < 4; i++ {
		items = append(items, ObsResidual{Residual: model.Residual{ResRA: 0.02, ResDec: 0.001, SepArcsec: 72}, StationID: "biased", CatalogID: "cat", Days: float64(i)})
		items = append(items, ObsResidual{Residual: model.Residual{ResRA: 0, ResDec: 0, SepArcsec: 0}, StationID: "nominal", CatalogID: "cat", Days: float64(i)})
	}
	a := Analyze("arc", items)
	if a.Kind != model.AttrKindStation || a.TargetID != "biased" {
		t.Fatalf("expected station/biased when station signal is stronger, got kind=%s target=%s", a.Kind, a.TargetID)
	}
}

func TestAnalyzePrefersStrongerCatalogOverStation(t *testing.T) {
	// 星表整体偏移且跨台站高度一致（两个台站都朝同方向偏移），台站组内仅有弱信号。
	// 此时应优先归因星表。
	items := make([]ObsResidual, 0, 8)
	for i := 0; i < 4; i++ {
		items = append(items, ObsResidual{Residual: model.Residual{ResRA: 0.02, ResDec: 0.01, SepArcsec: 80}, StationID: "stA", CatalogID: "cat", Days: float64(i)})
		items = append(items, ObsResidual{Residual: model.Residual{ResRA: 0.021, ResDec: 0.011, SepArcsec: 82}, StationID: "stB", CatalogID: "cat", Days: float64(i)})
	}
	a := Analyze("arc", items)
	if a.Kind != model.AttrKindCatalog || a.TargetID != "cat" {
		t.Fatalf("expected catalog/cat when catalog signal is stronger, got kind=%s target=%s", a.Kind, a.TargetID)
	}
}

func TestAnalyzeEmptyInputStaysCandidate(t *testing.T) {
	a := Analyze("arc", nil)
	if a.Kind != model.AttrKindCandidate || a.Status != model.AttrStatusCandidate || a.Confidence != 0 {
		t.Fatalf("empty input should remain a candidate: %+v", a)
	}
}
