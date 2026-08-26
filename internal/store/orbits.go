package store

import (
	"database/sql"
	"time"

	"task254-asterorbit/internal/model"
)

// CreateOrbit 写入候选/拟合轨道。
func (db *DB) CreateOrbit(o model.Orbit) error {
	if err := o.Validate(); err != nil {
		return err
	}
	_, err := db.Exec(`INSERT INTO orbits(id,arc_id,a,e,i_deg,om_deg,w_deg,m0_deg,epoch_jd,source,created_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
		o.ID, o.ArcID, o.A, o.E, o.IDeg, o.OmDeg, o.WDeg, o.M0Deg, o.EpochJD, o.Source,
		o.CreatedAt.UTC().Format(time.RFC3339))
	return err
}

// GetOrbit 读取单条轨道。
func (db *DB) GetOrbit(id string) (model.Orbit, error) {
	var o model.Orbit
	var ca string
	err := db.QueryRow(`SELECT id,arc_id,a,e,i_deg,om_deg,w_deg,m0_deg,epoch_jd,source,created_at
		FROM orbits WHERE id=?`, id).Scan(&o.ID, &o.ArcID, &o.A, &o.E, &o.IDeg, &o.OmDeg, &o.WDeg, &o.M0Deg, &o.EpochJD, &o.Source, &ca)
	if err == sql.ErrNoRows {
		return o, ErrNotFound
	}
	if err != nil {
		return o, err
	}
	o.CreatedAt = parseTime(ca)
	return o, nil
}

// GetOrbitByArc 读取弧段最新一条轨道（按创建时间）。
func (db *DB) GetOrbitByArc(arcID string) (model.Orbit, error) {
	var o model.Orbit
	var ca string
	err := db.QueryRow(`SELECT id,arc_id,a,e,i_deg,om_deg,w_deg,m0_deg,epoch_jd,source,created_at
		FROM orbits WHERE arc_id=? ORDER BY created_at DESC LIMIT 1`, arcID).
		Scan(&o.ID, &o.ArcID, &o.A, &o.E, &o.IDeg, &o.OmDeg, &o.WDeg, &o.M0Deg, &o.EpochJD, &o.Source, &ca)
	if err == sql.ErrNoRows {
		return o, ErrNotFound
	}
	if err != nil {
		return o, err
	}
	o.CreatedAt = parseTime(ca)
	return o, nil
}

// ListOrbitsByArc 列出弧段全部轨道。
func (db *DB) ListOrbitsByArc(arcID string) ([]model.Orbit, error) {
	rows, err := db.Query(`SELECT id,arc_id,a,e,i_deg,om_deg,w_deg,m0_deg,epoch_jd,source,created_at
		FROM orbits WHERE arc_id=? ORDER BY created_at ASC`, arcID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Orbit
	for rows.Next() {
		var o model.Orbit
		var ca string
		if err := rows.Scan(&o.ID, &o.ArcID, &o.A, &o.E, &o.IDeg, &o.OmDeg, &o.WDeg, &o.M0Deg, &o.EpochJD, &o.Source, &ca); err != nil {
			return nil, err
		}
		o.CreatedAt = parseTime(ca)
		out = append(out, o)
	}
	return out, rows.Err()
}

// UpsertResidual 删除该观测的旧残差后写入新残差（重算语义）。
func (db *DB) UpsertResidual(r model.Residual) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM residuals WHERE observation_id=?`, r.ObservationID); err != nil {
		return err
	}
	if _, err := tx.Exec(`INSERT INTO residuals(id,observation_id,arc_id,computed_ra_deg,computed_dec_deg,res_ra_deg,res_dec_deg,sep_arcsec,computed_at)
		VALUES(?,?,?,?,?,?,?,?,?)`,
		r.ID, r.ObservationID, r.ArcID, r.ComputedRA, r.ComputedDec, r.ResRA, r.ResDec, r.SepArcsec,
		r.ComputedAt.UTC().Format(time.RFC3339)); err != nil {
		return err
	}
	return tx.Commit()
}

// ListResidualsByArc 列出弧段全部残差。
func (db *DB) ListResidualsByArc(arcID string) ([]model.Residual, error) {
	rows, err := db.Query(`SELECT id,observation_id,arc_id,computed_ra_deg,computed_dec_deg,res_ra_deg,res_dec_deg,sep_arcsec,computed_at
		FROM residuals WHERE arc_id=? ORDER BY computed_at ASC`, arcID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Residual
	for rows.Next() {
		var r model.Residual
		var ca string
		if err := rows.Scan(&r.ID, &r.ObservationID, &r.ArcID, &r.ComputedRA, &r.ComputedDec, &r.ResRA, &r.ResDec, &r.SepArcsec, &ca); err != nil {
			return nil, err
		}
		r.ComputedAt = parseTime(ca)
		out = append(out, r)
	}
	return out, rows.Err()
}
