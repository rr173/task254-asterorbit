// 近地小行星轨道观测残差归因服务入口。
// 提供 HTTP 服务（--addr/--db）与 --smoke-test 自检契约。
package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"task254-asterorbit/internal/httpapi"
	"task254-asterorbit/internal/model"
	"task254-asterorbit/internal/propagation"
	"task254-asterorbit/internal/service"
	"task254-asterorbit/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP 监听地址")
	dbPath := flag.String("db", "asterorbit.db", "SQLite 数据库路径")
	smoke := flag.Bool("smoke-test", false, "运行端到端自检后立即退出（退出码 0 表示通过）")
	flag.Parse()

	if *smoke {
		if err := runSmokeTest(*dbPath); err != nil {
			fmt.Fprintln(os.Stderr, "smoke-test FAILED:", err)
			os.Exit(1)
		}
		fmt.Println("smoke-test OK")
		return
	}

	db, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	svc := service.New(db)
	r := httpapi.NewRouter(svc)
	log.Printf("asterorbit 监听 %s (db=%s)", *addr, *dbPath)
	if err := http.ListenAndServe(*addr, r); err != nil {
		log.Fatalf("http serve: %v", err)
	}
}

// runSmokeTest 真实建库、跑观测导入→传播→残差→归因→校准收敛→快照发布闭环，
// 关闭并重新打开数据库验证持久化与重启恢复，最后返回 nil 表示通过。
func runSmokeTest(dbPath string) error {
	path := dbPath
	if path == "" || path == "asterorbit.db" {
		path = filepath.Join(os.TempDir(), fmt.Sprintf("asterorbit_smoke_%d.db", os.Getpid()))
	}
	_ = os.Remove(path)
	defer os.Remove(path)

	db, err := store.Open(path)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	svc := service.New(db)

	if !svc.SelfCheck().OK {
		return fmt.Errorf("self-check failed")
	}

	cat, err := svc.CreateCatalog("GAIA-DR3", "J2000", "ICRS", 0, 0)
	if err != nil {
		return fmt.Errorf("catalog: %w", err)
	}
	// 台站 A 带未知零位偏差；台站 B 名义零位。
	staA, err := svc.CreateStation("Mount Bias", "BIA", 35.0, -120.0, 1700)
	if err != nil {
		return fmt.Errorf("station A: %w", err)
	}
	staB, err := svc.CreateStation("Mount Nominal", "NOM", -30.0, 20.0, 1000)
	if err != nil {
		return fmt.Errorf("station B: %w", err)
	}
	const (
		trueRAZeroArcsec  = 40.0 // 台站 A 真实未知赤经零位偏差（角秒）
		trueDecZeroArcsec = 0.0
	)
	arc, err := svc.CreateArc("2026 NEA 自检弧段", "2026 TEST")
	if err != nil {
		return fmt.Errorf("arc: %w", err)
	}
	epochJD := 2461000.5
	orb, err := svc.SetOrbit(arc.ID, 1.05, 0.20, 12.0, 75.0, 40.0, 5.0, epochJD, "smoke")
	if err != nil {
		return fmt.Errorf("orbit: %w", err)
	}
	el := propagation.OrbitElements{A: orb.A, E: orb.E, IDeg: orb.IDeg, OmDeg: orb.OmDeg, WDeg: orb.WDeg, M0Deg: orb.M0Deg, EpochJD: orb.EpochJD}

	base := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	nObs := 0
	for k := 0; k < 12; k++ {
		t := base.Add(time.Duration(k) * 3 * 24 * time.Hour)
		// 台站 A：观测值含真实零位偏差（合成时叠加）
		craA, cdecA := propagation.ComputeTopocentric(el, t, 35.0, -120.0, 1700, 0, 0, 0)
		if _, err := svc.ImportObservation(arc.ID, staA.ID, cat.ID, t, craA+trueRAZeroArcsec/3600.0, cdecA+trueDecZeroArcsec/3600.0, 0.3, 0.3); err != nil {
			return fmt.Errorf("import A: %w", err)
		}
		// 台站 B：名义零位
		craB, cdecB := propagation.ComputeTopocentric(el, t, -30.0, 20.0, 1000, 0, 0, 0)
		if _, err := svc.ImportObservation(arc.ID, staB.ID, cat.ID, t, craB, cdecB, 0.3, 0.3); err != nil {
			return fmt.Errorf("import B: %w", err)
		}
		nObs += 2
	}

	// 未校准残差计算
	n, err := svc.ComputeResiduals(arc.ID)
	if err != nil {
		return fmt.Errorf("compute: %w", err)
	}
	if n != nObs {
		return fmt.Errorf("expected %d residuals, got %d", nObs, n)
	}
	resBefore, err := svc.GetResiduals(arc.ID)
	if err != nil {
		return fmt.Errorf("residuals: %w", err)
	}
	rmsBefore := globalRMS(resBefore)

	// 归因：应判为台站 A 相关
	a0, err := svc.Analyze(arc.ID)
	if err != nil {
		return fmt.Errorf("analyze: %w", err)
	}
	if a0.Kind != model.AttrKindStation {
		return fmt.Errorf("expected station attribution, got %s", a0.Kind)
	}
	if a0.TargetID != staA.ID {
		return fmt.Errorf("expected biased station %s, got %s", staA.ID, a0.TargetID)
	}

	// 锁定台站 A 真实零位校准并重新计算
	if err := svc.ApplyCalibration(staA.ID, trueRAZeroArcsec, trueDecZeroArcsec, "smoke 恢复赤经零位"); err != nil {
		return fmt.Errorf("calibrate: %w", err)
	}
	if _, err := svc.ComputeResiduals(arc.ID); err != nil {
		return fmt.Errorf("recompute: %w", err)
	}
	resAfter, err := svc.GetResiduals(arc.ID)
	if err != nil {
		return fmt.Errorf("residuals after: %w", err)
	}
	rmsAfter := globalRMS(resAfter)
	if !(rmsAfter < rmsBefore*0.5) {
		return fmt.Errorf("calibration did not converge RMS: before=%.3f after=%.3f arcsec", rmsBefore, rmsAfter)
	}

	// 发布快照
	snap, err := svc.PublishOrbitSnapshot(arc.ID, cat.ID)
	if err != nil {
		return fmt.Errorf("publish: %w", err)
	}
	if snap.Hash == "" {
		return fmt.Errorf("snapshot hash empty")
	}

	// 重启恢复：关闭并重新打开数据库，验证持久化
	if err := db.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}
	db2, err := store.Open(path)
	if err != nil {
		return fmt.Errorf("reopen: %w", err)
	}
	defer db2.Close()
	svc2 := service.New(db2)
	arc2, err := svc2.GetArc(arc.ID)
	if err != nil {
		return fmt.Errorf("get arc after restart: %w", err)
	}
	if arc2.Status != model.ArcStatusPublished {
		return fmt.Errorf("arc status not persisted: %s", arc2.Status)
	}
	obs2, err := svc2.ListObservations(arc.ID)
	if err != nil {
		return fmt.Errorf("list obs after restart: %w", err)
	}
	if len(obs2) != nObs {
		return fmt.Errorf("observations not persisted: got %d", len(obs2))
	}
	snap2, err := svc2.GetLatestSnapshot(arc.ID)
	if err != nil {
		return fmt.Errorf("snapshot after restart: %w", err)
	}
	if snap2.Hash != snap.Hash {
		return fmt.Errorf("snapshot hash mismatch after restart")
	}
	return nil
}

func globalRMS(res []model.Residual) float64 {
	if len(res) == 0 {
		return 0
	}
	var ss float64
	for _, r := range res {
		ss += r.SepArcsec * r.SepArcsec
	}
	return math.Sqrt(ss / float64(len(res)))
}
