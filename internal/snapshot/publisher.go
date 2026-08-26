// Package snapshot 负责轨道快照的不可变发布：序列化轨道解、所采用星表与台站校准，
// 计算内容哈希并维护“发布→替代”的快照链。
package snapshot

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"task254-asterorbit/internal/model"
)

type payload struct {
	Orbit       model.Orbit     `json:"orbit"`
	CatalogID   string          `json:"catalog_id"`
	Stations    []model.Station `json:"stations"`
	PublishedAt string          `json:"published_at"`
}

// BuildPayload 把轨道解、星表与台站校准序列化为不可变快照内容（JSON）。
func BuildPayload(orb model.Orbit, catalogID string, stations []model.Station, publishedAt string) (string, error) {
	p := payload{Orbit: orb, CatalogID: catalogID, Stations: stations, PublishedAt: publishedAt}
	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ComputeHash 对快照内容计算 SHA-256 作为内容指纹。
func ComputeHash(content string) string {
	// PublishedAt is audit metadata, not scientific content.  Canonicalize it
	// away so retrying the same orbit/catalog/calibration snapshot is stable.
	var p payload
	if err := json.Unmarshal([]byte(content), &p); err == nil {
		p.PublishedAt = ""
		if canonical, err := json.Marshal(p); err == nil {
			content = string(canonical)
		}
	}
	h := sha256.Sum256([]byte(content))
	return hex.EncodeToString(h[:])
}
