package model

import "time"

// 残差来源种类与状态。
const (
	AttrKindCandidate = "candidate"
	AttrKindStation   = "station"
	AttrKindCatalog   = "catalog"
	AttrKindModel     = "model"

	AttrStatusCandidate = "candidate"
	AttrStatusConfirmed = "confirmed"
)

// Attribution 是对一段弧段残差来源的统计裁决：台站相关、星表相关或模型相关。
type Attribution struct {
	ID          string
	ArcID       string
	Kind        string
	TargetID    string // 台站或星表 ID；模型相关时为空
	Evidence    string
	MeanRADeg   float64
	MeanDecDeg  float64
	RMSArcsec   float64
	SlopePerDay float64
	Confidence  float64
	Status      string
	CreatedAt   time.Time
}

// Validate 校验归因实体。
func (a Attribution) Validate() error {
	switch a.Kind {
	case AttrKindCandidate, AttrKindStation, AttrKindCatalog, AttrKindModel:
		// ok
	default:
		return ErrInvalidArc
	}
	if a.Confidence < 0 || a.Confidence > 1 {
		return ErrInvalidArc
	}
	return nil
}
