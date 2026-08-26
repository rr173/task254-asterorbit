package store

import (
	"database/sql"
	"errors"
	"time"

	"task254-asterorbit/internal/model"
)

// ErrNotFound 表示查询实体不存在。
var ErrNotFound = errors.New("store: not found")

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

// CreateArc 写入观测弧段。
func (db *DB) CreateArc(a model.ObservationArc) error {
	if err := a.Validate(); err != nil {
		return err
	}
	_, err := db.Exec(`INSERT INTO arcs(id,name,status,object_name,created_at,updated_at)
		VALUES(?,?,?,?,?,?)`,
		a.ID, a.Name, a.Status, a.ObjectName,
		a.CreatedAt.UTC().Format(time.RFC3339), a.UpdatedAt.UTC().Format(time.RFC3339))
	return err
}

// GetArc 读取弧段。
func (db *DB) GetArc(id string) (model.ObservationArc, error) {
	var a model.ObservationArc
	var ca, ua string
	err := db.QueryRow(`SELECT id,name,status,object_name,created_at,updated_at FROM arcs WHERE id=?`, id).
		Scan(&a.ID, &a.Name, &a.Status, &a.ObjectName, &ca, &ua)
	if err == sql.ErrNoRows {
		return a, ErrNotFound
	}
	if err != nil {
		return a, err
	}
	a.CreatedAt = parseTime(ca)
	a.UpdatedAt = parseTime(ua)
	return a, nil
}

// ListArcs 列出全部弧段（按创建时间倒序）。
func (db *DB) ListArcs() ([]model.ObservationArc, error) {
	rows, err := db.Query(`SELECT id,name,status,object_name,created_at,updated_at FROM arcs ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.ObservationArc
	for rows.Next() {
		var a model.ObservationArc
		var ca, ua string
		if err := rows.Scan(&a.ID, &a.Name, &a.Status, &a.ObjectName, &ca, &ua); err != nil {
			return nil, err
		}
		a.CreatedAt, a.UpdatedAt = parseTime(ca), parseTime(ua)
		out = append(out, a)
	}
	return out, rows.Err()
}

// UpdateArcStatus 更新弧段状态（封存态禁止修改）。
func (db *DB) UpdateArcStatus(id, status string) error {
	a, err := db.GetArc(id)
	if err != nil {
		return err
	}
	if a.Sealed() {
		return model.ErrSealedArc
	}
	_, err = db.Exec(`UPDATE arcs SET status=?, updated_at=? WHERE id=?`, status, time.Now().UTC().Format(time.RFC3339), id)
	return err
}

// SealArc 封存弧段。
func (db *DB) SealArc(id string) error {
	return db.UpdateArcStatus(id, model.ArcStatusSealed)
}

// CreateObservation 写入测角观测；若 ID 已存在返回 ErrDuplicateObs（幂等保护）。
func (db *DB) CreateObservation(o model.Observation) error {
	if err := o.Validate(); err != nil {
		return err
	}
	if _, err := db.GetObservation(o.ID); err == nil {
		return model.ErrDuplicateObs
	}
	_, err := db.Exec(`INSERT INTO observations(id,arc_id,station_id,catalog_id,obs_time_utc,ra_deg,dec_deg,ra_err_arcsec,dec_err_arcsec,status,imported_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
		o.ID, o.ArcID, o.StationID, o.CatalogID, o.ObsTimeUTC.UTC().Format(time.RFC3339),
		o.RA, o.Dec, o.RAErrArcsec, o.DecErrArcsec, o.Status, o.ImportedAt.UTC().Format(time.RFC3339))
	return err
}

// GetObservation 读取单条观测。
func (db *DB) GetObservation(id string) (model.Observation, error) {
	var o model.Observation
	var ot, ia string
	err := db.QueryRow(`SELECT id,arc_id,station_id,catalog_id,obs_time_utc,ra_deg,dec_deg,ra_err_arcsec,dec_err_arcsec,status,imported_at
		FROM observations WHERE id=?`, id).
		Scan(&o.ID, &o.ArcID, &o.StationID, &o.CatalogID, &ot, &o.RA, &o.Dec, &o.RAErrArcsec, &o.DecErrArcsec, &o.Status, &ia)
	if err == sql.ErrNoRows {
		return o, ErrNotFound
	}
	if err != nil {
		return o, err
	}
	o.ObsTimeUTC = parseTime(ot)
	o.ImportedAt = parseTime(ia)
	return o, nil
}

// ListObservationsByArc 按弧段列出观测（按观测时刻升序）。
func (db *DB) ListObservationsByArc(arcID string) ([]model.Observation, error) {
	rows, err := db.Query(`SELECT id,arc_id,station_id,catalog_id,obs_time_utc,ra_deg,dec_deg,ra_err_arcsec,dec_err_arcsec,status,imported_at
		FROM observations WHERE arc_id=? ORDER BY obs_time_utc ASC`, arcID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Observation
	for rows.Next() {
		var o model.Observation
		var ot, ia string
		if err := rows.Scan(&o.ID, &o.ArcID, &o.StationID, &o.CatalogID, &ot, &o.RA, &o.Dec, &o.RAErrArcsec, &o.DecErrArcsec, &o.Status, &ia); err != nil {
			return nil, err
		}
		o.ObsTimeUTC = parseTime(ot)
		o.ImportedAt = parseTime(ia)
		out = append(out, o)
	}
	return out, rows.Err()
}

// UpdateObservationStatus 更新观测状态（排除/异常等）。
func (db *DB) UpdateObservationStatus(id, status string) error {
	_, err := db.Exec(`UPDATE observations SET status=? WHERE id=?`, status, id)
	return err
}

// CountObservationsByArc 统计弧段观测数。
func (db *DB) CountObservationsByArc(arcID string) (int, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM observations WHERE arc_id=?`, arcID).Scan(&n)
	return n, err
}
