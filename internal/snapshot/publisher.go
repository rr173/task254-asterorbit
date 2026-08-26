// Package snapshot 负责轨道快照的不可变发布：序列化轨道解、所采用星表与台站校准，
// 计算内容哈希并维护“发布→替代”的快照链。
package snapshot

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"

	"task254-asterorbit/internal/model"
)

// payload 是快照的完整不可变内容：科学字段加上发布元数据（published_at）。
// published_at 属于发布元数据而非科学内容，因此不参与内容指纹计算——
// 否则重复发布同一轨道解会因为时间戳不同而产生不同指纹，无法识别为同一份可复现结果。
type payload struct {
	Orbit       model.Orbit     `json:"orbit"`
	CatalogID   string          `json:"catalog_id"`
	Stations    []model.Station `json:"stations"`
	PublishedAt string         `json:"published_at"`
}

// scientificContent 仅包含决定实验可复现性的科学字段（轨道根数+星表版本+台站校准），
// 不含任何发布元数据，是内容指纹的稳定输入。
type scientificContent struct {
	Orbit     model.Orbit     `json:"orbit"`
	CatalogID string          `json:"catalog_id"`
	Stations  []model.Station `json:"stations"`
}

// BuildPayload 把轨道解、星表与台站校准序列化为不可变快照内容（JSON）。
// 返回的字符串仍含 published_at（供审计展示），但该字段不参与内容指纹。
func BuildPayload(orb model.Orbit, catalogID string, stations []model.Station, publishedAt string) (string, error) {
	p := payload{Orbit: orb, CatalogID: catalogID, Stations: stations, PublishedAt: publishedAt}
	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// BuildContent 把轨道解、星表与台站校准序列化为参与内容指纹的科学内容（JSON）。
// 仅含科学字段，不含 published_at 等发布元数据，保证相同科学内容产生稳定指纹。
// 台站作为集合参与指纹，故按 ID 规范化排序，使指纹不依赖底层存储的返回顺序。
func BuildContent(orb model.Orbit, catalogID string, stations []model.Station) (string, error) {
	sorted := make([]model.Station, len(stations))
	copy(sorted, stations)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ID < sorted[j].ID })
	c := scientificContent{Orbit: orb, CatalogID: catalogID, Stations: sorted}
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ComputeHash 对快照内容计算 SHA-256 作为内容指纹。
func ComputeHash(content string) string {
	h := sha256.Sum256([]byte(content))
	return hex.EncodeToString(h[:])
}
