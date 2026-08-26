package model

import "time"

// Station 是测角观测台站。RAZeroArcsec / DecZeroArcsec 为台站零位校准量
// （赤经/赤纬系统偏差，单位角秒），在残差归因判定为台站相关后由研究者锁定并写回，
// 作为对计算位置的加性修正，使残差收敛。
type Station struct {
	ID           string
	Name         string
	Code         string
	LatitudeDeg  float64
	LongitudeDeg float64
	HeightM      float64
	RAZeroArcsec float64
	DecZeroArcsec float64
	CreatedAt    time.Time
}

// Validate 校验台站坐标与代码。
func (s Station) Validate() error {
	if s.ID == "" || s.Code == "" || s.Name == "" {
		return ErrInvalidStation
	}
	if s.LatitudeDeg < -90 || s.LatitudeDeg > 90 {
		return ErrInvalidStation
	}
	if s.LongitudeDeg < -180 || s.LongitudeDeg > 180 {
		return ErrInvalidStation
	}
	if s.HeightM < -500 || s.HeightM > 9000 {
		return ErrInvalidStation
	}
	return nil
}
