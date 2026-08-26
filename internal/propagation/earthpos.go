package propagation

import "math"

// EarthEclipticJ2000 用 Meeus 低精度太阳历表（1990 年代精度约 0.01°）得到地球日心位置，
// 再把平黄道-of-date 经度经岁差约化到 J2000 黄道，返回日心 J2000 黄道直角坐标（AU）。
func EarthEclipticJ2000(jd float64) Vec3 {
	T := (jd - 2451545.0) / 36525.0
	L0 := 280.46646 + 36000.76983*T + 0.0003032*T*T
	M := 357.52911 + 35999.05029*T - 0.0001537*T*T
	e := 0.016708634 - 0.000042037*T - 0.0000001267*T*T
	Mr := M * D2R
	C := (1.914602-0.004817*T-0.000014*T*T)*math.Sin(Mr) +
		(0.019993-0.000101*T)*math.Sin(2*Mr) + 0.000289*math.Sin(3*Mr)
	trueLong := (L0 + C) * D2R
	R := (1.000001018 * (1 - e*e)) / (1 + e*math.Cos(trueLong-Mr))
	// 地球日心黄道经度 = 太阳地心黄经 + 180°
	lamE := trueLong + math.Pi
	xe := R * math.Cos(lamE)
	ye := R * math.Sin(lamE)
	// 岁差绕黄极旋转 -p（p 为自 J2000 起的黄经岁差，单位度）
	p := (50.2879 / 3600.0) * T * D2R
	x := xe*math.Cos(p) - ye*math.Sin(p)
	y := xe*math.Sin(p) + ye*math.Cos(p)
	return Vec3{X: x, Y: y, Z: 0}
}
