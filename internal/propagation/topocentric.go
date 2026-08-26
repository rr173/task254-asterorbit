package propagation

import (
	"math"
	"time"

	"task254-asterorbit/internal/timescale"
)

// GMSTDeg 由 UTC 儒略日计算 Greenwich 平恒星时（度，IAU 1982 公式）。
func GMSTDeg(jdUTC float64) float64 {
	T := (jdUTC - 2451545.0) / 36525.0
	sec := 67310.54841 + (876600*3600+8640184.812866)*T + 0.093104*T*T - 6.2e-6*T*T*T
	sec = math.Mod(sec, 86400)
	if sec < 0 {
		sec += 86400
	}
	return sec / 240.0
}

// StationGeocentricECI 由台站大地坐标（WGS84）计算其地心赤道惯性（J2000）位置，单位 AU。
func StationGeocentricECI(latDeg, lonDeg, heightM, jdUTC float64) Vec3 {
	lat := latDeg * D2R
	lon := lonDeg * D2R
	a := 6378.137
	f := 1 / 298.257223563
	e2 := f * (2 - f)
	N := a / math.Sqrt(1-e2*math.Sin(lat)*math.Sin(lat))
	h := heightM / 1000.0
	xEcef := (N + h) * math.Cos(lat) * math.Cos(lon)
	yEcef := (N + h) * math.Cos(lat) * math.Sin(lon)
	zEcef := (N*(1-e2) + h) * math.Sin(lat)
	gst := GMSTDeg(jdUTC) * D2R
	xEci := xEcef*math.Cos(gst) - yEcef*math.Sin(gst)
	yEci := xEcef*math.Sin(gst) + yEcef*math.Cos(gst)
	return Vec3{X: xEci / AUKM, Y: yEci / AUKM, Z: zEcef / AUKM}
}

// ComputeTopocentric 由轨道根数、观测 UTC 时刻与台站坐标计算站心赤经赤纬（度）。
// clockOffsetS、latBiasArcsec、longBiasArcsec 为台站校准量：钟差改观测历元，
// 经纬度偏差修正站址；零校准量时即“未校准计算位置”，其残差可暴露台站系统偏差。
func ComputeTopocentric(el OrbitElements, obsUTC time.Time, latDeg, lonDeg, heightM, clockOffsetS, latBiasArcsec, longBiasArcsec float64) (raDeg, decDeg float64) {
	effUTC := obsUTC.Add(time.Duration(clockOffsetS * 1e9))
	jdTDB := timescale.TimeToJDTDB(effUTC)
	ast := Propagate(el, jdTDB)
	earth := EarthEclipticJ2000(jdTDB)
	geo := Vec3{X: ast.X - earth.X, Y: ast.Y - earth.Y, Z: ast.Z - earth.Z}
	eq := eclipticToEquatorialJ2000(geo)
	latC := latDeg + latBiasArcsec/3600.0
	lonC := lonDeg + longBiasArcsec/3600.0
	sta := StationGeocentricECI(latC, lonC, heightM, timescale.JDFromTime(effUTC))
	topo := Vec3{X: eq.X - sta.X, Y: eq.Y - sta.Y, Z: eq.Z - sta.Z}
	r := math.Sqrt(topo.X*topo.X + topo.Y*topo.Y + topo.Z*topo.Z)
	ra := math.Atan2(topo.Y, topo.X) / D2R
	if ra < 0 {
		ra += 360
	}
	dec := math.Asin(topo.Z/r) / D2R
	return ra, dec
}

// AngularSepArcsec 计算两个天球点 (ra,dec) 的角距（角秒）。
func AngularSepArcsec(ra1, dec1, ra2, dec2 float64) float64 {
	r1, d1 := ra1*D2R, dec1*D2R
	r2, d2 := ra2*D2R, dec2*D2R
	cosSep := math.Sin(d1)*math.Sin(d2) + math.Cos(d1)*math.Cos(d2)*math.Cos(r1-r2)
	if cosSep > 1 {
		cosSep = 1
	}
	if cosSep < -1 {
		cosSep = -1
	}
	return math.Acos(cosSep) / D2R * 3600.0
}
