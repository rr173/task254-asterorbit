package model

import "time"

// 观测弧段状态机：接收中 → 待拟合 → 需复核 → 已发布 → 封存。
const (
	ArcStatusReceiving   = "receiving"
	ArcStatusPendingFit  = "pending_fit"
	ArcStatusNeedsReview = "needs_review"
	ArcStatusPublished   = "published"
	ArcStatusSealed      = "sealed"
)

// ObservationArc 是一段围绕同一小行星的测角观测集合，是归因分析的工作单元。
type ObservationArc struct {
	ID         string
	Name       string
	Status     string
	ObjectName string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Validate 校验弧段实体。
func (a ObservationArc) Validate() error {
	if a.ID == "" || a.Name == "" || a.ObjectName == "" {
		return ErrInvalidArc
	}
	switch a.Status {
	case ArcStatusReceiving, ArcStatusPendingFit, ArcStatusNeedsReview, ArcStatusPublished, ArcStatusSealed:
		return nil
	}
	return ErrInvalidArc
}

// Sealed 表示弧段已封存，禁止修改。
func (a ObservationArc) Sealed() bool { return a.Status == ArcStatusSealed }

// CanAdvanceArc reports whether a lifecycle transition is allowed.  Repeating
// the current state is harmless and keeps retries idempotent.
func CanAdvanceArc(from, to string) bool {
	if from == to {
		return from == ArcStatusReceiving || from == ArcStatusPendingFit || from == ArcStatusNeedsReview || from == ArcStatusPublished || from == ArcStatusSealed
	}
	switch from {
	case ArcStatusReceiving:
		return to == ArcStatusPendingFit
	case ArcStatusPendingFit:
		return to == ArcStatusNeedsReview
	case ArcStatusNeedsReview:
		return to == ArcStatusPublished
	case ArcStatusPublished:
		return to == ArcStatusSealed
	default:
		return false
	}
}
