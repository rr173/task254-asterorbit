package service

import (
	"time"

	"task254-asterorbit/internal/attribution"
	"task254-asterorbit/internal/model"
	"task254-asterorbit/internal/observation"
	"task254-asterorbit/internal/snapshot"
)

// anomalyThresholdArcsec 残差超过该角距判为异常观测，待研究者复核。
const anomalyThresholdArcsec = 30.0

// ImportObservation 导入一条测角观测，幂等（同观测重复导入不新建），
// 弧段自动由接收中推进到待拟合。
func (s *Service) ImportObservation(arcID, stationID, catalogID string, obsUTC time.Time, ra, dec, raErr, decErr float64) (model.Observation, error) {
	arc, err := s.DB.GetArc(arcID)
	if err != nil {
		return model.Observation{}, err
	}
	if arc.Sealed() {
		return model.Observation{}, model.ErrSealedArc
	}
	if _, err := s.DB.GetStation(stationID); err != nil {
		return model.Observation{}, err
	}
	if _, err := s.DB.GetCatalog(catalogID); err != nil {
		return model.Observation{}, err
	}
	id := observation.DeriveObservationID(arcID, stationID, catalogID, obsUTC, ra, dec)
	o := model.Observation{
		ID: id, ArcID: arcID, StationID: stationID, CatalogID: catalogID,
		ObsTimeUTC: obsUTC, RA: ra, Dec: dec, RAErrArcsec: raErr, DecErrArcsec: decErr,
		Status: model.ObsStatusRaw, ImportedAt: now(),
	}
	if err := s.DB.CreateObservation(o); err != nil {
		if err == model.ErrDuplicateObs {
			return s.DB.GetObservation(id) // 幂等：返回已存在记录
		}
		return model.Observation{}, err
	}
	if arc.Status == model.ArcStatusReceiving {
		_ = s.DB.UpdateArcStatus(arcID, model.ArcStatusPendingFit)
	}
	return o, nil
}

// ComputeResiduals 用当前弧段轨道与台站校准重算全部观测的 O-C 残差。
func (s *Service) ComputeResiduals(arcID string) (int, error) {
	orb, err := s.DB.GetOrbitByArc(arcID)
	if err != nil {
		return 0, model.ErrArcNotFitted
	}
	obs, err := s.DB.ListObservationsByArc(arcID)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, o := range obs {
		if o.Status == model.ObsStatusExcluded {
			continue
		}
		st, err := s.DB.GetStation(o.StationID)
		if err != nil {
			return n, err
		}
		r := observation.ComputeResidual(o, orb, st)
		r.ID = "res_" + o.ID
		r.ComputedAt = now()
		if err := s.DB.UpsertResidual(r); err != nil {
			return n, err
		}
		n++
		if observation.FlagAnomalous(r.SepArcsec, anomalyThresholdArcsec) {
			_ = s.DB.UpdateObservationStatus(o.ID, model.ObsStatusAnomalous)
		}
	}
	return n, nil
}

// Analyze 对弧段残差做来源裁决并落库，弧段推进到需复核。
func (s *Service) Analyze(arcID string) (model.Attribution, error) {
	residuals, err := s.DB.ListResidualsByArc(arcID)
	if err != nil {
		return model.Attribution{}, err
	}
	allObs, err := s.DB.ListObservationsByArc(arcID)
	if err != nil {
		return model.Attribution{}, err
	}
	obsMap := make(map[string]model.Observation, len(allObs))
	var t0 time.Time
	for _, o := range allObs {
		obsMap[o.ID] = o
		if t0.IsZero() || o.ObsTimeUTC.After(t0) {
			t0 = o.ObsTimeUTC
		}
	}
	items := make([]attribution.ObsResidual, 0, len(residuals))
	for _, r := range residuals {
		o := obsMap[r.ObservationID]
		days := t0.Sub(o.ObsTimeUTC).Hours() / 24.0
		items = append(items, attribution.ObsResidual{
			Residual:  r,
			StationID: o.StationID,
			CatalogID: o.CatalogID,
			Days:      days,
		})
	}
	a := attribution.Analyze(arcID, items)
	a.ID = newID("attr")
	a.CreatedAt = now()
	if err := s.DB.CreateAttribution(a); err != nil {
		return a, err
	}
	_ = s.DB.UpdateArcStatus(arcID, model.ArcStatusNeedsReview)
	return a, nil
}

// ApplyCalibration 锁定台站零位校准量（赤经/赤纬，角秒），写回台站并留历史，
// 用于修正计算位置使残差收敛。
func (s *Service) ApplyCalibration(stationID string, raZero, decZero float64, note string) error {
	if _, err := s.DB.GetStation(stationID); err != nil {
		return err
	}
	return s.DB.ApplyStationCalibration(stationID, newID("cal"), raZero, decZero, note)
}

// PublishOrbitSnapshot 发布当前轨道解快照（冻结星表与台站校准，维护替代链），弧段发布。
func (s *Service) PublishOrbitSnapshot(arcID, catalogID string) (model.OrbitSnapshot, error) {
	orb, err := s.DB.GetOrbitByArc(arcID)
	if err != nil {
		return model.OrbitSnapshot{}, err
	}
	catID := catalogID
	if catID == "" {
		cats, _ := s.DB.ListCatalogs()
		if len(cats) > 0 {
			catID = cats[0].ID
		}
	}
	stations, err := s.DB.ListStations()
	if err != nil {
		return model.OrbitSnapshot{}, err
	}
	publishedAt := now().Format(time.RFC3339)
	payload, err := snapshot.BuildPayload(orb, catID, stations, publishedAt)
	if err != nil {
		return model.OrbitSnapshot{}, err
	}
	snap := model.OrbitSnapshot{
		ID:        newID("snap"),
		ArcID:     arcID,
		OrbitID:   orb.ID,
		CatalogID: catID,
		Hash:      snapshot.ComputeHash(payload),
		Status:    model.SnapStatusPublished,
		CreatedAt: now(),
		Payload:   payload,
	}
	if err := s.DB.PublishSnapshot(snap); err != nil {
		return model.OrbitSnapshot{}, err
	}
	_ = s.DB.UpdateArcStatus(arcID, model.ArcStatusPublished)
	return snap, nil
}
