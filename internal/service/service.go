// Package service 是业务编排层，串接 store 与各算法包，对 httpapi 暴露高层操作。
package service

import (
	"crypto/rand"
	"encoding/hex"
	"math"
	"time"

	"task254-asterorbit/internal/model"
	"task254-asterorbit/internal/propagation"
	"task254-asterorbit/internal/store"
)

// now 返回当前 UTC 时间（统一时间戳来源）。
func now() time.Time { return time.Now().UTC() }

func newID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return prefix + "_" + hex.EncodeToString(b)
}

// Service 持有数据库句柄，提供高层业务操作。
type Service struct{ DB *store.DB }

// New 构造 Service。
func New(db *store.DB) *Service { return &Service{DB: db} }

// CreateArc 新建观测弧段（初始为接收中）。
func (s *Service) CreateArc(name, objectName string) (model.ObservationArc, error) {
	a := model.ObservationArc{
		ID:         newID("arc"),
		Name:       name,
		Status:     model.ArcStatusReceiving,
		ObjectName: objectName,
		CreatedAt:  now(),
		UpdatedAt:  now(),
	}
	if err := s.DB.CreateArc(a); err != nil {
		return model.ObservationArc{}, err
	}
	return a, nil
}

// GetArc / ListArcs / AdvanceArc 弧段查询与状态推进。
func (s *Service) GetArc(id string) (model.ObservationArc, error) { return s.DB.GetArc(id) }
func (s *Service) ListArcs() ([]model.ObservationArc, error)       { return s.DB.ListArcs() }
func (s *Service) AdvanceArc(id, status string) error             { return s.DB.UpdateArcStatus(id, status) }
func (s *Service) SealArc(id string) error                        { return s.DB.SealArc(id) }

// CreateStation / GetStation / ListStations 台站管理。
func (s *Service) CreateStation(name, code string, lat, lon, h float64) (model.Station, error) {
	st := model.Station{ID: newID("sta"), Name: name, Code: code, LatitudeDeg: lat, LongitudeDeg: lon, HeightM: h, CreatedAt: now()}
	if err := s.DB.CreateStation(st); err != nil {
		return model.Station{}, err
	}
	return st, nil
}
func (s *Service) GetStation(id string) (model.Station, error)   { return s.DB.GetStation(id) }
func (s *Service) ListStations() ([]model.Station, error)        { return s.DB.ListStations() }
func (s *Service) StationByCode(code string) (model.Station, error) { return s.DB.StationByCode(code) }

// CreateCatalog / GetCatalog / ListCatalogs 参考星表管理。
func (s *Service) CreateCatalog(name, epoch, refFrame string, biasRA, biasDec float64) (model.StarCatalog, error) {
	c := model.StarCatalog{ID: newID("cat"), Name: name, Epoch: epoch, RefFrame: refFrame, BiasRAArcsec: biasRA, BiasDecArcsec: biasDec, CreatedAt: now()}
	if err := s.DB.CreateCatalog(c); err != nil {
		return model.StarCatalog{}, err
	}
	return c, nil
}
func (s *Service) GetCatalog(id string) (model.StarCatalog, error) { return s.DB.GetCatalog(id) }
func (s *Service) ListCatalogs() ([]model.StarCatalog, error)      { return s.DB.ListCatalogs() }

// SetOrbit 为弧段写入轨道根数（候选/拟合解）。
func (s *Service) SetOrbit(arcID string, a, e, iDeg, omDeg, wDeg, m0Deg, epochJD float64, source string) (model.Orbit, error) {
	o := model.Orbit{ID: newID("orb"), ArcID: arcID, A: a, E: e, IDeg: iDeg, OmDeg: omDeg, WDeg: wDeg, M0Deg: m0Deg, EpochJD: epochJD, Source: source, CreatedAt: now()}
	if err := s.DB.CreateOrbit(o); err != nil {
		return model.Orbit{}, err
	}
	return o, nil
}
func (s *Service) GetOrbit(arcID string) (model.Orbit, error)  { return s.DB.GetOrbitByArc(arcID) }
func (s *Service) ListOrbits(arcID string) ([]model.Orbit, error) { return s.DB.ListOrbitsByArc(arcID) }

// ListObservations / GetResiduals / GetAttribution / ListAttributions / GetSnapshot / ListSnapshots 查询。
func (s *Service) ListObservations(arcID string) ([]model.Observation, error) { return s.DB.ListObservationsByArc(arcID) }
func (s *Service) GetResiduals(arcID string) ([]model.Residual, error)        { return s.DB.ListResidualsByArc(arcID) }
func (s *Service) GetAttribution(arcID string) (model.Attribution, error)     { return s.DB.GetLatestAttribution(arcID) }
func (s *Service) ListAttributions(arcID string) ([]model.Attribution, error) { return s.DB.ListAttributions(arcID) }
func (s *Service) GetSnapshot(id string) (model.OrbitSnapshot, error)          { return s.DB.GetSnapshot(id) }
func (s *Service) GetLatestSnapshot(arcID string) (model.OrbitSnapshot, error) { return s.DB.GetLatestPublishedSnapshot(arcID) }
func (s *Service) ListSnapshots(arcID string) ([]model.OrbitSnapshot, error)   { return s.DB.ListSnapshots(arcID) }

// SelfCheckResult 自检结果。
type SelfCheckResult struct {
	OK                bool     `json:"ok"`
	DBPing            bool     `json:"db_ping"`
	TableCount        int      `json:"table_count"`
	PropagationSanity bool     `json:"propagation_sanity"`
	Notes             []string `json:"notes"`
}

// SelfCheck 校验数据库可达、表齐全与轨道传播数值 sanity。
func (s *Service) SelfCheck() SelfCheckResult {
	res := SelfCheckResult{Notes: []string{}}
	if err := s.DB.Ping(); err != nil {
		res.Notes = append(res.Notes, "db ping failed: "+err.Error())
		return res
	}
	res.DBPing = true
	_ = s.DB.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table'`).Scan(&res.TableCount)
	// 传播 sanity：圆轨道半径应恒等于 a。
	el := propagation.OrbitElements{A: 1.0, E: 0, IDeg: 0, OmDeg: 0, WDeg: 0, M0Deg: 0, EpochJD: 2451545.0}
	p := propagation.Propagate(el, 2451545.0+12.345)
	rad := math.Hypot(p.X, math.Hypot(p.Y, p.Z))
	res.PropagationSanity = math.Abs(rad-1.0) < 1e-6
	res.OK = res.DBPing && res.PropagationSanity
	if !res.PropagationSanity {
		res.Notes = append(res.Notes, "propagation radius sanity off")
	}
	return res
}
