// Package observation 负责测角观测的导入幂等化、校验与残差计算。
package observation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"task254-asterorbit/internal/model"
)

// DeriveObservationID 由 (弧段, 台站, 星表, 时刻, 赤经, 赤纬) 生成确定性幂等 ID，
// 保证同一观测重复导入不会创建两条记录。
func DeriveObservationID(arcID, stationID, catalogID string, t time.Time, ra, dec float64) string {
	raw := fmt.Sprintf("%s|%s|%s|%s|%.6f|%.6f",
		arcID, stationID, catalogID, t.UTC().Format(time.RFC3339Nano), ra, dec)
	h := sha256.Sum256([]byte(raw))
	return "obs_" + hex.EncodeToString(h[:])[:16]
}

// ValidateObservation 校验测角记录业务合法性。
func ValidateObservation(o model.Observation) error {
	return o.Validate()
}

// FlagAnomalous 依据角距阈值判定观测是否异常（残差过大视为异常，待研究者复核）。
func FlagAnomalous(sepArcsec, thresholdArcsec float64) bool {
	return sepArcsec > thresholdArcsec
}
