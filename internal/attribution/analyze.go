// Package attribution 对一段弧段的 O-C 残差做模式分析，裁决其来源：
// 台站相关（校准）、星表相关（参考星表系统偏差）或模型相关（未建模动力学）。
package attribution

import (
	"fmt"
	"math"

	"task254-asterorbit/internal/model"
)

// ObsResidual 把残差与其来源上下文（台站、星表、相对首观测的天数）绑定，供模式分析。
type ObsResidual struct {
	model.Residual
	StationID string
	CatalogID string
	Days      float64
}

type groupStat struct {
	count     int
	meanRA    float64 // 度
	meanDec   float64 // 度
	offsetArc float64 // 平均残差向量模 角秒
	rmsArc    float64 // 组内 RMS 角秒
}

func summarize(items []ObsResidual, key func(ObsResidual) string) map[string]groupStat {
	out := map[string]groupStat{}
	for _, it := range items {
		k := key(it)
		g := out[k]
		g.count++
		g.meanRA += it.ResRA
		g.meanDec += it.ResDec
		out[k] = g
	}
	for k, g := range out {
		g.meanRA /= float64(g.count)
		g.meanDec /= float64(g.count)
		var ss float64
		for _, it := range items {
			if key(it) != k {
				continue
			}
			dra := (it.ResRA - g.meanRA) * 3600
			ddc := (it.ResDec - g.meanDec) * 3600
			ss += dra*dra + ddc*ddc
		}
		g.rmsArc = math.Sqrt(ss / float64(g.count))
		g.offsetArc = math.Hypot(g.meanRA, g.meanDec) * 3600
		out[k] = g
	}
	return out
}

func globalRMS(items []ObsResidual) float64 {
	if len(items) == 0 {
		return 0
	}
	var ss float64
	for _, it := range items {
		dra := it.ResRA * 3600
		ddc := it.ResDec * 3600
		ss += dra*dra + ddc*ddc
	}
	return math.Sqrt(ss / float64(len(items)))
}

// linregSlope 对 (x,y) 做最小二乘斜率，返回斜率与皮尔逊相关系数。
func linregSlope(xs, ys []float64) (slope, corr float64) {
	n := float64(len(xs))
	if n < 2 {
		return 0, 0
	}
	var sx, sy, sxx, sxy, syy float64
	for i := range xs {
		sx += xs[i]
		sy += ys[i]
		sxx += xs[i] * xs[i]
		sxy += xs[i] * ys[i]
		syy += ys[i] * ys[i]
	}
	denom := n*sxx - sx*sx
	if denom == 0 {
		return 0, 0
	}
	slope = (n*sxy - sx*sy) / denom
	var num float64
	num = n*sxy - sx*sy
	cov := num
	vx := n*sxx - sx*sx
	vy := n*syy - sy*sy
	if vx <= 0 || vy <= 0 {
		return slope, 0
	}
	corr = cov / math.Sqrt(vx*vy)
	return slope, corr
}

// Analyze 对弧段残差做来源裁决，返回一条候选归因（Kind 可能为 candidate）。
func Analyze(arcID string, items []ObsResidual) model.Attribution {
	a := model.Attribution{
		ArcID:    arcID,
		Kind:     model.AttrKindCandidate,
		Status:   model.AttrStatusCandidate,
		Evidence: "样本不足或无明显系统性模式，留待研究者复核",
	}
	if len(items) == 0 {
		return a
	}
	gRMS := globalRMS(items)
	byStation := summarize(items, func(it ObsResidual) string { return it.StationID })
	byCatalog := summarize(items, func(it ObsResidual) string { return it.CatalogID })

	// 台站信号：某台站残差平均偏移较大且组内高度一致（方向稳定、组内 RMS 远小于偏移量）。
	bestStation, bestStationStat, bestStationSig := "", groupStat{}, 0.0
	for id, g := range byStation {
		if g.count < 2 {
			continue
		}
		sig := g.offsetArc / math.Max(g.rmsArc, 1e-3)
		if g.offsetArc > 5 && g.rmsArc < 0.5*g.offsetArc && sig > 2 && sig > bestStationSig {
			bestStation, bestStationStat, bestStationSig = id, g, sig
		}
	}
	// 星表信号：某星表残差平均偏移跨台站一致地偏大且组内一致。
	bestCatalog, bestCatalogStat, bestCatalogSig := "", groupStat{}, 0.0
	for id, g := range byCatalog {
		if g.count < 2 {
			continue
		}
		sig := g.offsetArc / math.Max(g.rmsArc, 1e-3)
		if g.offsetArc > 5 && g.rmsArc < 0.5*g.offsetArc && sig > 1.5 && sig > bestCatalogSig {
			bestCatalog, bestCatalogStat, bestCatalogSig = id, g, sig
		}
	}
	// 模型信号：角距随观测天数单调增长（未建模扰动累积）。
	var xs, ys []float64
	for _, it := range items {
		xs = append(xs, it.Days)
		ys = append(ys, it.SepArcsec)
	}
	slope, corr := linregSlope(xs, ys)

	switch {
	case bestStation != "" && bestCatalog == "":
		a.Kind = model.AttrKindStation
		a.TargetID = bestStation
		a.MeanRADeg = bestStationStat.meanRA
		a.MeanDecDeg = bestStationStat.meanDec
		a.RMSArcsec = bestStationStat.rmsArc
		a.Confidence = clamp(bestStationSig / (bestStationSig + 1))
		a.Evidence = fmt.Sprintf("台站 %s 残差持续朝固定方向偏移 %.2f 角秒（组内 RMS %.2f 角秒，信噪比 %.1f），符合台站钟差/站址偏差特征",
			bestStation, bestStationStat.offsetArc, bestStationStat.rmsArc, bestStationSig)
	case bestCatalog != "":
		a.Kind = model.AttrKindCatalog
		a.TargetID = bestCatalog
		a.MeanRADeg = bestCatalogStat.meanRA
		a.MeanDecDeg = bestCatalogStat.meanDec
		a.RMSArcsec = bestCatalogStat.rmsArc
		a.Confidence = clamp(bestCatalogSig / (bestCatalogSig + 1))
		a.Evidence = fmt.Sprintf("星表 %s 残差整体偏移 %.2f 角秒（信噪比 %.1f），跨台站一致，符合参考星表系统偏差特征",
			bestCatalog, bestCatalogStat.offsetArc, bestCatalogSig)
	case slope > 0.05 && corr > 0.4:
		a.Kind = model.AttrKindModel
		a.SlopePerDay = slope
		a.RMSArcsec = gRMS
		a.Confidence = clamp(corr)
		a.Evidence = fmt.Sprintf("残差角距随观测天数以 %.4f 角秒/天增长（相关系数 %.2f），符合未建模动力学（辐射压/行星摄动）特征", slope, corr)
	default:
		a.RMSArcsec = gRMS
		a.Evidence = fmt.Sprintf("全局 RMS %.2f 角秒，未检出显著台站/星表/模型模式", gRMS)
	}
	return a
}

func clamp(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}
