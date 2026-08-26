package model

import "time"

// Residual 是单条观测的观测减计算（O-C）残差，单位为度；SepArcsec 为角距（角秒）。
type Residual struct {
	ID           string
	ObservationID string
	ArcID        string
	ComputedRA   float64
	ComputedDec  float64
	ResRA        float64
	ResDec       float64
	SepArcsec    float64
	ComputedAt   time.Time
}
