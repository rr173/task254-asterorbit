// Package model 定义近地小行星轨道观测残差归因服务的核心实体、状态常量与领域错误。
package model

import "errors"

// 领域错误。
var (
	ErrInvalidArc     = errors.New("asterorbit: invalid observation arc")
	ErrInvalidObs     = errors.New("asterorbit: invalid observation")
	ErrInvalidStation = errors.New("asterorbit: invalid station")
	ErrInvalidOrbit   = errors.New("asterorbit: invalid orbit elements")
	ErrInvalidCatalog = errors.New("asterorbit: invalid star catalog")
	ErrSealedArc      = errors.New("asterorbit: arc is sealed, modification refused")
	ErrFrozenSnapshot = errors.New("asterorbit: snapshot is frozen, modification refused")
	ErrDuplicateObs   = errors.New("asterorbit: duplicate observation id")
	ErrArcNotFitted   = errors.New("asterorbit: arc has no fitted orbit")
	ErrBadCoord       = errors.New("asterorbit: ra/dec out of range")
)

// 赤经范围 [0,360) 度，赤纬范围 [-90,90] 度。
func validSkyCoord(ra, dec float64) bool {
	return ra >= 0 && ra < 360 && dec >= -90 && dec <= 90
}
