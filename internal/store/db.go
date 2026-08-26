// Package store 提供基于 modernc.org/sqlite（纯 Go、CGO 无关）的持久化层，
// 建表迁移与各实体 CRUD，支持关闭重开后的重启恢复。
package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// DB 封装 *sql.DB，现代码库统一通过它访问 SQLite。
type DB struct{ *sql.DB }

// Open 打开（必要时创建）SQLite 数据库并完成迁移。
func Open(path string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("store: open %s: %w", path, err)
	}
	sqlDB.SetMaxOpenConns(1) // SQLite 串行写，避免写锁竞争
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("store: ping: %w", err)
	}
	db := &DB{sqlDB}
	if err := db.Migrate(); err != nil {
		return nil, err
	}
	return db, nil
}

// Migrate 建表（幂等）。所有实体均有主键，重启后数据完整保留。
func (db *DB) Migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS star_catalogs (
			id TEXT PRIMARY KEY, name TEXT NOT NULL, epoch TEXT, ref_frame TEXT,
			bias_ra_arcsec REAL, bias_dec_arcsec REAL, created_at TEXT)`,
		`CREATE TABLE IF NOT EXISTS stations (
			id TEXT PRIMARY KEY, name TEXT NOT NULL, code TEXT UNIQUE NOT NULL,
			latitude_deg REAL, longitude_deg REAL, height_m REAL,
			ra_zero_arcsec REAL, dec_zero_arcsec REAL, created_at TEXT)`,
		`CREATE TABLE IF NOT EXISTS arcs (
			id TEXT PRIMARY KEY, name TEXT NOT NULL, status TEXT NOT NULL,
			object_name TEXT NOT NULL, created_at TEXT, updated_at TEXT)`,
		`CREATE TABLE IF NOT EXISTS observations (
			id TEXT PRIMARY KEY, arc_id TEXT NOT NULL, station_id TEXT NOT NULL, catalog_id TEXT NOT NULL,
			obs_time_utc TEXT NOT NULL, ra_deg REAL, dec_deg REAL, ra_err_arcsec REAL, dec_err_arcsec REAL,
			status TEXT NOT NULL, imported_at TEXT)`,
		`CREATE TABLE IF NOT EXISTS calibrations (
			id TEXT PRIMARY KEY, station_id TEXT NOT NULL, ra_zero_arcsec REAL,
			dec_zero_arcsec REAL, applied_at TEXT, note TEXT)`,
		`CREATE TABLE IF NOT EXISTS orbits (
			id TEXT PRIMARY KEY, arc_id TEXT NOT NULL, a REAL, e REAL, i_deg REAL, om_deg REAL,
			w_deg REAL, m0_deg REAL, epoch_jd REAL, source TEXT, created_at TEXT)`,
		`CREATE TABLE IF NOT EXISTS residuals (
			id TEXT PRIMARY KEY, observation_id TEXT NOT NULL, arc_id TEXT NOT NULL,
			computed_ra_deg REAL, computed_dec_deg REAL, res_ra_deg REAL, res_dec_deg REAL,
			sep_arcsec REAL, computed_at TEXT)`,
		`CREATE TABLE IF NOT EXISTS attributions (
			id TEXT PRIMARY KEY, arc_id TEXT NOT NULL, kind TEXT NOT NULL, target_id TEXT,
			evidence TEXT, mean_ra_deg REAL, mean_dec_deg REAL, rms_arcsec REAL, slope_per_day REAL,
			confidence REAL, status TEXT NOT NULL, created_at TEXT)`,
		`CREATE TABLE IF NOT EXISTS snapshots (
			id TEXT PRIMARY KEY, arc_id TEXT NOT NULL, orbit_id TEXT NOT NULL, catalog_id TEXT,
			hash TEXT NOT NULL, status TEXT NOT NULL, supersedes TEXT, created_at TEXT, payload TEXT)`,
		`CREATE INDEX IF NOT EXISTS idx_obs_arc ON observations(arc_id)`,
		`CREATE INDEX IF NOT EXISTS idx_res_arc ON residuals(arc_id)`,
		`CREATE INDEX IF NOT EXISTS idx_attr_arc ON attributions(arc_id)`,
		`CREATE INDEX IF NOT EXISTS idx_snap_arc ON snapshots(arc_id)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("store: migrate: %w", err)
		}
	}
	return nil
}
