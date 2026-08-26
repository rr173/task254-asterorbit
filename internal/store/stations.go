package store

import (
	"database/sql"
	"time"

	"task254-asterorbit/internal/model"
)

// CreateStation 写入台站。
func (db *DB) CreateStation(s model.Station) error {
	if err := s.Validate(); err != nil {
		return err
	}
	_, err := db.Exec(`INSERT INTO stations(id,name,code,latitude_deg,longitude_deg,height_m,ra_zero_arcsec,dec_zero_arcsec,created_at)
		VALUES(?,?,?,?,?,?,?,?,?)`,
		s.ID, s.Name, s.Code, s.LatitudeDeg, s.LongitudeDeg, s.HeightM,
		s.RAZeroArcsec, s.DecZeroArcsec, s.CreatedAt.UTC().Format(time.RFC3339))
	return err
}

// GetStation 读取台站。
func (db *DB) GetStation(id string) (model.Station, error) {
	var s model.Station
	var ca string
	err := db.QueryRow(`SELECT id,name,code,latitude_deg,longitude_deg,height_m,ra_zero_arcsec,dec_zero_arcsec,created_at
		FROM stations WHERE id=?`, id).Scan(&s.ID, &s.Name, &s.Code, &s.LatitudeDeg, &s.LongitudeDeg, &s.HeightM,
		&s.RAZeroArcsec, &s.DecZeroArcsec, &ca)
	if err == sql.ErrNoRows {
		return s, ErrNotFound
	}
	s.CreatedAt = parseTime(ca)
	return s, err
}

// StationByCode 按台站代码读取。
func (db *DB) StationByCode(code string) (model.Station, error) {
	var s model.Station
	var ca string
	err := db.QueryRow(`SELECT id,name,code,latitude_deg,longitude_deg,height_m,ra_zero_arcsec,dec_zero_arcsec,created_at
		FROM stations WHERE code=?`, code).Scan(&s.ID, &s.Name, &s.Code, &s.LatitudeDeg, &s.LongitudeDeg, &s.HeightM,
		&s.RAZeroArcsec, &s.DecZeroArcsec, &ca)
	if err == sql.ErrNoRows {
		return s, ErrNotFound
	}
	s.CreatedAt = parseTime(ca)
	return s, err
}

// ListStations 列出全部台站。
func (db *DB) ListStations() ([]model.Station, error) {
	rows, err := db.Query(`SELECT id,name,code,latitude_deg,longitude_deg,height_m,ra_zero_arcsec,dec_zero_arcsec,created_at FROM stations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Station
	for rows.Next() {
		var s model.Station
		var ca string
		if err := rows.Scan(&s.ID, &s.Name, &s.Code, &s.LatitudeDeg, &s.LongitudeDeg, &s.HeightM,
			&s.RAZeroArcsec, &s.DecZeroArcsec, &ca); err != nil {
			return nil, err
		}
		s.CreatedAt = parseTime(ca)
		out = append(out, s)
	}
	return out, rows.Err()
}

// ApplyStationCalibration 把台站零位校准量写回台站当前字段（修正计算位置），并记录校准历史。
func (db *DB) ApplyStationCalibration(stationID, calibID string, raZero, decZero float64, note string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`UPDATE stations SET ra_zero_arcsec=?, dec_zero_arcsec=? WHERE id=?`,
		raZero, decZero, stationID); err != nil {
		return err
	}
	if _, err := tx.Exec(`INSERT INTO calibrations(id,station_id,ra_zero_arcsec,dec_zero_arcsec,applied_at,note)
		VALUES(?,?,?,?,?,?)`,
		calibID, stationID, raZero, decZero, time.Now().UTC().Format(time.RFC3339), note); err != nil {
		return err
	}
	return tx.Commit()
}

// CreateCatalog 写入参考星表。
func (db *DB) CreateCatalog(c model.StarCatalog) error {
	if err := c.Validate(); err != nil {
		return err
	}
	_, err := db.Exec(`INSERT INTO star_catalogs(id,name,epoch,ref_frame,bias_ra_arcsec,bias_dec_arcsec,created_at)
		VALUES(?,?,?,?,?,?,?)`,
		c.ID, c.Name, c.Epoch, c.RefFrame, c.BiasRAArcsec, c.BiasDecArcsec, c.CreatedAt.UTC().Format(time.RFC3339))
	return err
}

// GetCatalog 读取星表。
func (db *DB) GetCatalog(id string) (model.StarCatalog, error) {
	var c model.StarCatalog
	var ca string
	err := db.QueryRow(`SELECT id,name,epoch,ref_frame,bias_ra_arcsec,bias_dec_arcsec,created_at FROM star_catalogs WHERE id=?`, id).
		Scan(&c.ID, &c.Name, &c.Epoch, &c.RefFrame, &c.BiasRAArcsec, &c.BiasDecArcsec, &ca)
	if err == sql.ErrNoRows {
		return c, ErrNotFound
	}
	c.CreatedAt = parseTime(ca)
	return c, err
}

// ListCatalogs 列出星表。
func (db *DB) ListCatalogs() ([]model.StarCatalog, error) {
	rows, err := db.Query(`SELECT id,name,epoch,ref_frame,bias_ra_arcsec,bias_dec_arcsec,created_at FROM star_catalogs`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.StarCatalog
	for rows.Next() {
		var c model.StarCatalog
		var ca string
		if err := rows.Scan(&c.ID, &c.Name, &c.Epoch, &c.RefFrame, &c.BiasRAArcsec, &c.BiasDecArcsec, &ca); err != nil {
			return nil, err
		}
		c.CreatedAt = parseTime(ca)
		out = append(out, c)
	}
	return out, rows.Err()
}

// CalibrationRecord 是台站校准历史记录。
type CalibrationRecord struct {
	ID            string
	StationID     string
	RAZeroArcsec  float64
	DecZeroArcsec float64
	AppliedAt     time.Time
	Note          string
}

// ListCalibrationsByStation 列出某台站校准历史（按应用时间升序）。
func (db *DB) ListCalibrationsByStation(stationID string) ([]CalibrationRecord, error) {
	rows, err := db.Query(`SELECT id,station_id,ra_zero_arcsec,dec_zero_arcsec,applied_at,note
		FROM calibrations WHERE station_id=? ORDER BY applied_at ASC`, stationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CalibrationRecord
	for rows.Next() {
		var c CalibrationRecord
		var at string
		if err := rows.Scan(&c.ID, &c.StationID, &c.RAZeroArcsec, &c.DecZeroArcsec, &at, &c.Note); err != nil {
			return nil, err
		}
		c.AppliedAt = parseTime(at)
		out = append(out, c)
	}
	return out, rows.Err()
}
