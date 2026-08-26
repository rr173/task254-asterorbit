package store

import (
	"database/sql"
	"time"

	"task254-asterorbit/internal/model"
)

// CreateAttribution 写入残差来源裁决。
func (db *DB) CreateAttribution(a model.Attribution) error {
	if err := a.Validate(); err != nil {
		return err
	}
	_, err := db.Exec(`INSERT INTO attributions(id,arc_id,kind,target_id,evidence,mean_ra_deg,mean_dec_deg,rms_arcsec,slope_per_day,confidence,status,created_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
		a.ID, a.ArcID, a.Kind, a.TargetID, a.Evidence, a.MeanRADeg, a.MeanDecDeg, a.RMSArcsec, a.SlopePerDay, a.Confidence, a.Status,
		a.CreatedAt.UTC().Format(time.RFC3339))
	return err
}

// GetLatestAttribution 读取弧段最新一条裁决。
func (db *DB) GetLatestAttribution(arcID string) (model.Attribution, error) {
	var a model.Attribution
	var ca string
	err := db.QueryRow(`SELECT id,arc_id,kind,target_id,evidence,mean_ra_deg,mean_dec_deg,rms_arcsec,slope_per_day,confidence,status,created_at
		FROM attributions WHERE arc_id=? ORDER BY created_at DESC LIMIT 1`, arcID).
		Scan(&a.ID, &a.ArcID, &a.Kind, &a.TargetID, &a.Evidence, &a.MeanRADeg, &a.MeanDecDeg, &a.RMSArcsec, &a.SlopePerDay, &a.Confidence, &a.Status, &ca)
	if err == sql.ErrNoRows {
		return a, ErrNotFound
	}
	if err != nil {
		return a, err
	}
	a.CreatedAt = parseTime(ca)
	return a, nil
}

// ListAttributions 列出弧段全部裁决。
func (db *DB) ListAttributions(arcID string) ([]model.Attribution, error) {
	rows, err := db.Query(`SELECT id,arc_id,kind,target_id,evidence,mean_ra_deg,mean_dec_deg,rms_arcsec,slope_per_day,confidence,status,created_at
		FROM attributions WHERE arc_id=? ORDER BY created_at ASC`, arcID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Attribution
	for rows.Next() {
		var a model.Attribution
		var ca string
		if err := rows.Scan(&a.ID, &a.ArcID, &a.Kind, &a.TargetID, &a.Evidence, &a.MeanRADeg, &a.MeanDecDeg, &a.RMSArcsec, &a.SlopePerDay, &a.Confidence, &a.Status, &ca); err != nil {
			return nil, err
		}
		a.CreatedAt = parseTime(ca)
		out = append(out, a)
	}
	return out, rows.Err()
}

// CreateSnapshot 写入轨道快照。
func (db *DB) CreateSnapshot(s model.OrbitSnapshot) error {
	_, err := db.Exec(`INSERT INTO snapshots(id,arc_id,orbit_id,catalog_id,hash,status,supersedes,created_at,payload)
		VALUES(?,?,?,?,?,?,?,?,?)`,
		s.ID, s.ArcID, s.OrbitID, s.CatalogID, s.Hash, s.Status, s.Supersedes, s.CreatedAt.UTC().Format(time.RFC3339), s.Payload)
	return err
}

// PublishSnapshot 发布快照并把上一版已发布快照标记为替代（维护发布→替代链）。
func (db *DB) PublishSnapshot(s model.OrbitSnapshot) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var prevID string
	_ = tx.QueryRow(`SELECT id FROM snapshots WHERE arc_id=? AND status=? ORDER BY created_at DESC LIMIT 1`,
		s.ArcID, model.SnapStatusPublished).Scan(&prevID)
	if prevID != "" {
		if _, err := tx.Exec(`UPDATE snapshots SET status=? WHERE id=?`, model.SnapStatusSuperseded, prevID); err != nil {
			return err
		}
		s.Supersedes = prevID
	}
	s.Status = model.SnapStatusPublished
	if _, err := tx.Exec(`INSERT INTO snapshots(id,arc_id,orbit_id,catalog_id,hash,status,supersedes,created_at,payload)
		VALUES(?,?,?,?,?,?,?,?,?)`,
		s.ID, s.ArcID, s.OrbitID, s.CatalogID, s.Hash, s.Status, s.Supersedes, s.CreatedAt.UTC().Format(time.RFC3339), s.Payload); err != nil {
		return err
	}
	return tx.Commit()
}

// GetLatestPublishedSnapshot 读取弧段当前已发布快照。
func (db *DB) GetLatestPublishedSnapshot(arcID string) (model.OrbitSnapshot, error) {
	var s model.OrbitSnapshot
	var ca string
	err := db.QueryRow(`SELECT id,arc_id,orbit_id,catalog_id,hash,status,supersedes,created_at,payload
		FROM snapshots WHERE arc_id=? AND status=? ORDER BY created_at DESC LIMIT 1`, arcID, model.SnapStatusPublished).
		Scan(&s.ID, &s.ArcID, &s.OrbitID, &s.CatalogID, &s.Hash, &s.Status, &s.Supersedes, &ca, &s.Payload)
	if err == sql.ErrNoRows {
		return s, ErrNotFound
	}
	if err != nil {
		return s, err
	}
	s.CreatedAt = parseTime(ca)
	return s, nil
}

// GetSnapshot 读取快照。
func (db *DB) GetSnapshot(id string) (model.OrbitSnapshot, error) {
	var s model.OrbitSnapshot
	var ca string
	err := db.QueryRow(`SELECT id,arc_id,orbit_id,catalog_id,hash,status,supersedes,created_at,payload
		FROM snapshots WHERE id=?`, id).
		Scan(&s.ID, &s.ArcID, &s.OrbitID, &s.CatalogID, &s.Hash, &s.Status, &s.Supersedes, &ca, &s.Payload)
	if err == sql.ErrNoRows {
		return s, ErrNotFound
	}
	if err != nil {
		return s, err
	}
	s.CreatedAt = parseTime(ca)
	return s, nil
}

// ListSnapshots 列出弧段全部快照。
func (db *DB) ListSnapshots(arcID string) ([]model.OrbitSnapshot, error) {
	rows, err := db.Query(`SELECT id,arc_id,orbit_id,catalog_id,hash,status,supersedes,created_at,payload
		FROM snapshots WHERE arc_id=? ORDER BY created_at ASC`, arcID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.OrbitSnapshot
	for rows.Next() {
		var s model.OrbitSnapshot
		var ca string
		if err := rows.Scan(&s.ID, &s.ArcID, &s.OrbitID, &s.CatalogID, &s.Hash, &s.Status, &s.Supersedes, &ca, &s.Payload); err != nil {
			return nil, err
		}
		s.CreatedAt = parseTime(ca)
		out = append(out, s)
	}
	return out, rows.Err()
}
