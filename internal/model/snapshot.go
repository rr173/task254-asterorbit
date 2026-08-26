package model

import "time"

// 轨道快照状态机：草稿 → 发布 → 替代。
const (
	SnapStatusDraft      = "draft"
	SnapStatusPublished  = "published"
	SnapStatusSuperseded = "superseded"
)

// OrbitSnapshot 是发布时冻结的轨道解（含所采用星表版本与台站校准），不可变且带内容哈希。
type OrbitSnapshot struct {
	ID         string
	ArcID      string
	OrbitID    string
	CatalogID  string
	Hash       string
	Status     string
	Supersedes string
	CreatedAt  time.Time
	Payload    string // 轨道根数+校准量的 JSON 快照
}
