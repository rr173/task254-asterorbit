package model

import "time"

// StarCatalog 是测角观测所采用的参考星表版本，自带系统偏差估计字段。
type StarCatalog struct {
	ID            string
	Name          string
	Epoch         string
	RefFrame      string
	BiasRAArcsec  float64
	BiasDecArcsec float64
	CreatedAt     time.Time
}

// Validate 校验星表实体完整性。
func (c StarCatalog) Validate() error {
	if c.ID == "" || c.Name == "" {
		return ErrInvalidCatalog
	}
	if c.BiasRAArcsec < -600 || c.BiasRAArcsec > 600 {
		return ErrInvalidCatalog
	}
	if c.BiasDecArcsec < -600 || c.BiasDecArcsec > 600 {
		return ErrInvalidCatalog
	}
	return nil
}
