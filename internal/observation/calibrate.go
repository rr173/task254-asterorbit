package observation

import (
	"math"

	"task254-asterorbit/internal/model"
	"task254-asterorbit/internal/propagation"
)

// toElements 把持久化轨道根数转为传播所需结构。
func toElements(orb model.Orbit) propagation.OrbitElements {
	return propagation.OrbitElements{
		A:       orb.A,
		E:       orb.E,
		IDeg:    orb.IDeg,
		OmDeg:   orb.OmDeg,
		WDeg:    orb.WDeg,
		M0Deg:   orb.M0Deg,
		EpochJD: orb.EpochJD,
	}
}

// wrapResidualRA 把赤经残差（度）归一化到 [-180,180)，
// 避免 0°/360° 边界附近被误判为相差近一整圈：
// 例如观测赤经≈0°、计算赤经≈360° 时，朴素相减得≈-360°，
// 实际应取≈0° 的跨边界短残差。
func wrapResidualRA(resRADeg float64) float64 {
	res := math.Mod(resRADeg, 360.0)
	if res >= 180.0 {
		res -= 360.0
	}
	if res < -180.0 {
		res += 360.0
	}
	return res
}

// ComputeResidual 计算单条观测的 O-C 残差（单位：度）。
// 在几何传播得到的计算位置上叠加台站零位校准量（赤经/赤纬角秒），作为对计算位置的加性修正；
// 零校准时即“未校准残差”，其系统性偏移正是归因分析要识别的台站信号。
func ComputeResidual(o model.Observation, orb model.Orbit, st model.Station) model.Residual {
	el := toElements(orb)
	cra, cdec := propagation.ComputeTopocentric(el, o.ObsTimeUTC,
		st.LatitudeDeg, st.LongitudeDeg, st.HeightM, 0, 0, 0)
	cra += st.RAZeroArcsec / 3600.0
	cdec += st.DecZeroArcsec / 3600.0
	// 赤经是周期量，需按跨 0°/360° 边界取最短有符号残差，否则会在边界处
	// 把近重合的观测/计算位置误判为相差近一圈，污染异常标记与归因统计。
	resRA := wrapResidualRA(o.RA - cra)
	resDec := o.Dec - cdec
	sep := propagation.AngularSepArcsec(o.RA, o.Dec, cra, cdec)
	return model.Residual{
		ObservationID: o.ID,
		ArcID:         o.ArcID,
		ComputedRA:    cra,
		ComputedDec:   cdec,
		ResRA:         resRA,
		ResDec:        resDec,
		SepArcsec:     sep,
	}
}
