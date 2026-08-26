package model

import "time"

// 测角记录状态机：原始 → 已校准 → 异常 → 排除。
const (
	ObsStatusRaw       = "raw"
	ObsStatusCalibrated = "calibrated"
	ObsStatusAnomalous = "anomalous"
	ObsStatusExcluded  = "excluded"
)

// Observation 是一条台站对某小行星在某个 UTC 时刻测得的赤经赤纬。
type Observation struct {
	ID           string
	ArcID        string
	StationID    string
	CatalogID    string
	ObsTimeUTC   time.Time
	RA           float64
	Dec          float64
	RAErrArcsec  float64
	DecErrArcsec float64
	Status       string
	ImportedAt   time.Time
}

// Validate 校验测角记录坐标与状态。
func (o Observation) Validate() error {
	if o.ID == "" || o.ArcID == "" || o.StationID == "" || o.CatalogID == "" {
		return ErrInvalidObs
	}
	if !validSkyCoord(o.RA, o.Dec) {
		return ErrBadCoord
	}
	switch o.Status {
	case ObsStatusRaw, ObsStatusCalibrated, ObsStatusAnomalous, ObsStatusExcluded:
		return nil
	}
	return ErrInvalidObs
}
