// Package propagation 实现二体开普勒轨道传播、低精度地球历表与站心几何，
// 把轨道根数转成可观测的站心赤经赤纬。
package propagation

import "math"

const (
	// KGAUSS 高斯引力常数，单位 AU^1.5 / day。
	KGAUSS = 0.01720209895
	// D2R 角度转弧度。
	D2R = math.Pi / 180.0
	// ObliquityJ2000 J2000 黄赤交角。
	ObliquityJ2000 = 23.4392911 * D2R
	// AUKM 一天文单位对应的公里数。
	AUKM = 149597870.7
)

// Vec3 是三维直角坐标向量。
type Vec3 struct{ X, Y, Z float64 }

// OrbitElements 是某 TDB 历元的开普勒轨道根数（J2000 黄道坐标系，角度单位为度）。
type OrbitElements struct {
	A      float64 // 半长轴 AU
	E      float64 // 偏心率
	IDeg   float64 // 轨道倾角 度
	OmDeg  float64 // 升交点黄经 度
	WDeg   float64 // 近心点幅角 度
	M0Deg  float64 // 历元平近点角 度
	EpochJD float64 // 历元儒略日（TDB）
}

// solveKepler 牛顿迭代求解开普勒方程 M = E - e sinE。
func solveKepler(M, e float64) float64 {
	M = math.Mod(M, 2*math.Pi)
	if M < 0 {
		M += 2 * math.Pi
	}
	E := M
	if e > 0.8 {
		E = math.Pi
	}
	for i := 0; i < 200; i++ {
		f := E - e*math.Sin(E) - M
		fp := 1 - e*math.Cos(E)
		d := f / fp
		E -= d
		if math.Abs(d) < 1e-13 {
			break
		}
	}
	return E
}

// Propagate 把轨道根数传播到目标 TDB 历元 jd，返回日心 J2000 黄道直角坐标（AU）。
func Propagate(el OrbitElements, jd float64) Vec3 {
	n := KGAUSS / math.Pow(el.A, 1.5) // rad/day
	M := el.M0Deg*D2R + n*(jd-el.EpochJD)
	E := solveKepler(M, el.E)
	nu := 2*math.Atan2(math.Sqrt(1+el.E)*math.Sin(E/2), math.Sqrt(1-el.E)*math.Cos(E/2))
	r := el.A * (1 - el.E*math.Cos(E))
	xp := r * math.Cos(nu)
	yp := r * math.Sin(nu)
	return perifocalToEcliptic(xp, yp, el.IDeg*D2R, el.OmDeg*D2R, el.WDeg*D2R)
}

// perifocalToEcliptic 把近焦平面坐标旋转到 J2000 黄道坐标系。
func perifocalToEcliptic(xp, yp, inc, om, w float64) Vec3 {
	cosO, sinO := math.Cos(om), math.Sin(om)
	cosI, sinI := math.Cos(inc), math.Sin(inc)
	cosW, sinW := math.Cos(w), math.Sin(w)
	x := (cosO*cosW - sinO*sinW*cosI)*xp + (-cosO*sinW - sinO*cosW*cosI)*yp
	y := (sinO*cosW + cosO*sinW*cosI)*xp + (-sinO*sinW + cosO*cosW*cosI)*yp
	z := (sinW*sinI)*xp + (cosW*sinI)*yp
	return Vec3{X: x, Y: y, Z: z}
}

// eclipticToEquatorialJ2000 把黄道直角坐标旋转到 J2000 赤道坐标系。
func eclipticToEquatorialJ2000(v Vec3) Vec3 {
	c, s := math.Cos(ObliquityJ2000), math.Sin(ObliquityJ2000)
	return Vec3{
		X: v.X,
		Y: v.Y*c - v.Z*s,
		Z: v.Y*s + v.Z*c,
	}
}
