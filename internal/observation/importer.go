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
//
// 时刻统一按 UTC 绝对时间点参与哈希：同一物理时刻无论以哪个时区文字表示
// （如 `Z`、`+09:00`、`-05:00`）都得到同一 ID，从而保证重复导入幂等；
// 而真正不同的时刻因时间点不同而得到不同 ID，不会互相覆盖。
// 这里刻意不使用 t.Format(time.RFC3339Nano) 这类带时区偏移的字符串——
// 那会让同一时刻因时区表示不同而派生不同 ID，导致重复导入新建记录。
func DeriveObservationID(arcID, stationID, catalogID string, t time.Time, ra, dec float64) string {
	raw := fmt.Sprintf("%s|%s|%s|%d|%.6f|%.6f",
		arcID, stationID, catalogID, t.UTC().UnixNano(), ra, dec)
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
