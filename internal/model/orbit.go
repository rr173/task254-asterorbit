package model

import "time"

// Orbit 是某历元的开普勒轨道根数（J2000 黄道坐标系，二体近似）。
type Orbit struct {
	ID       string
	ArcID    string
	A        float64 // 半长轴 AU
	E        float64 // 偏心率
	IDeg     float64 // 轨道倾角 度
	OmDeg    float64 // 升交点黄经 度
	WDeg     float64 // 近心点幅角 度
	M0Deg    float64 // 历元平近点角 度
	EpochJD  float64 // 历元儒略日（TDB）
	Source   string
	CreatedAt time.Time
}

// Validate 校验轨道根数物理合理性与取值范围。
func (o Orbit) Validate() error {
	if o.ID == "" || o.ArcID == "" {
		return ErrInvalidOrbit
	}
	if o.A <= 0 || o.A > 1000 {
		return ErrInvalidOrbit
	}
	if o.E < 0 || o.E >= 1 {
		return ErrInvalidOrbit
	}
	if o.IDeg < 0 || o.IDeg > 180 {
		return ErrInvalidOrbit
	}
	if o.EpochJD < 2400000 || o.EpochJD > 2600000 {
		return ErrInvalidOrbit
	}
	return nil
}
