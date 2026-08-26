package observation

import (
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

// ComputeResidual 计算单条观测的 O-C 残差（单位：度）。
// 在几何传播得到的计算位置上叠加台站零位校准量（赤经/赤纬角秒），作为对计算位置的加性修正；
// 零校准时即“未校准残差”，其系统性偏移正是归因分析要识别的台站信号。
func ComputeResidual(o model.Observation, orb model.Orbit, st model.Station) model.Residual {
	el := toElements(orb)
	cra, cdec := propagation.ComputeTopocentric(el, o.ObsTimeUTC,
		st.LatitudeDeg, st.LongitudeDeg, st.HeightM, 0, 0, 0)
	cra -= st.RAZeroArcsec / 3600.0
	cdec += st.DecZeroArcsec / 3600.0
	resRA := o.RA - cra
	resDec := o.Dec - cdec
	// RA 残差折回 [-180,180] 度便于统计
	if resRA > 180 {
		resRA -= 360
	}
	if resRA < -180 {
		resRA += 360
	}
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
