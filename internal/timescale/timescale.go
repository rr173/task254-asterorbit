// Package timescale 负责天文时间尺度统一：UTC → TAI → TT → TDB，
// 为轨道传播提供连续的动力学时（TDB）历元。
package timescale

import (
	"math"
	"time"
)

// JDFromUnix 由 Unix 秒与纳秒返回 UTC 儒略日。
func JDFromUnix(sec int64, nsec int) float64 {
	return float64(sec)/86400.0 + float64(nsec)/1e9/86400.0 + 2440587.5
}

// JDFromTime 返回给定时刻（按其时区解释）的 UTC 儒略日。
func JDFromTime(t time.Time) float64 {
	return JDFromUnix(t.Unix(), t.Nanosecond())
}

// UTCToTAI 在 UTC 上叠加跳秒得到 TAI。
func UTCToTAI(utc time.Time) time.Time {
	return utc.Add(time.Duration(TAIOffsetSeconds(utc)) * time.Second)
}

// TAIToTT TAI 与 TT 固定相差 32.184 秒。
func TAIToTT(tai time.Time) time.Time {
	return tai.Add(32184 * time.Millisecond)
}

// UTCToTT UTC → TT。
func UTCToTT(utc time.Time) time.Time {
	return TAIToTT(UTCToTAI(utc))
}

// TTToTDB 在 TT 儒略日上叠加周期项得到 TDB（Fairhead & Bretagnon 低精度近似），返回 TDB 儒略日。
func TTToTDB(ttJD float64) float64 {
	T := (ttJD - 2451545.0) / 36525.0
	g1 := math.Mod(628.3076*T+6.2401, 2*math.Pi)
	g2 := math.Mod(1256.6152*T+4.7191, 2*math.Pi)
	g3 := math.Mod(18849.227*T+1.394, 2*math.Pi)
	dtDays := 0.001658*math.Sin(g1) + 0.0000142*math.Sin(g2) + 0.0000001*math.Sin(g3)
	return ttJD + dtDays/86400.0
}

// TimeToJDTDB 把 UTC 时刻转换为 TDB 儒略日（轨道传播的动力学时历元）。
func TimeToJDTDB(utc time.Time) float64 {
	return TTToTDB(JDFromTime(UTCToTT(utc)))
}
