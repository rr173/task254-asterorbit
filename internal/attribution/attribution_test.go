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

// TestAnalyzeDetectsModelTrendFromGrowingResidual 验证残差随观测天数稳定增长时，
// 应给出动力学模型相关归因，且斜率符号为正、量级与时间跨度一致（回归任务核心修复点）。
func TestAnalyzeDetectsModelTrendFromGrowingResidual(t *testing.T) {
	// Days 相对首观测单调递增；SepArcsec 随天数线性增长，斜率 0.5 角秒/天。
	const slopePerDay = 0.5
	const spanDays = 60.0
	items := make([]ObsResidual, 0, 7)
	for k := 0; k <= 6; k++ {
		days := spanDays * float64(k) / 6.0
		items = append(items, ObsResidual{
			Residual:  model.Residual{SepArcsec: slopePerDay * days},
			StationID: "sta", CatalogID: "cat", Days: days,
		})
	}
	a := Analyze("arc", items)
	if a.Kind != model.AttrKindModel {
		t.Fatalf("growing residual should be attributed to model dynamics, got kind=%s evidence=%s", a.Kind, a.Evidence)
	}
	if a.SlopePerDay <= 0 {
		t.Fatalf("model slope should be positive (residual grows with time), got %.6f", a.SlopePerDay)
	}
	// 斜率量级应与实际增长速率一致，而非被符号/跨度错配扭曲。
	if diff := a.SlopePerDay - slopePerDay; diff > 0.05 || diff < -0.05 {
		t.Fatalf("model slope %.4f does not match true rate %.4f arcsec/day", a.SlopePerDay, slopePerDay)
	}
	// 所报速率外推到整个跨度应与残差累计量级一致。
	if got := a.SlopePerDay * spanDays; got < slopePerDay*spanDays*0.9 || got > slopePerDay*spanDays*1.1 {
		t.Fatalf("slope*span %.2f does not match expected cumulative %.2f", got, slopePerDay*spanDays)
	}
}
