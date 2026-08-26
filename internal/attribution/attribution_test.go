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

func TestAnalyzeEmptyInputStaysCandidate(t *testing.T) {
	a := Analyze("arc", nil)
	if a.Kind != model.AttrKindCandidate || a.Status != model.AttrStatusCandidate || a.Confidence != 0 {
		t.Fatalf("empty input should remain a candidate: %+v", a)
	}
}
