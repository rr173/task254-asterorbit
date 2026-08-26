package httpapi

import (
	"errors"
	"net/http"
	"time"

	"task254-asterorbit/internal/store"
)

func (rt *Router) handleCreateArc(w http.ResponseWriter, r *http.Request, _ []string) {
	var b struct {
		Name       string `json:"name"`
		ObjectName string `json:"object_name"`
	}
	if err := readJSON(r, &b); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	arc, err := rt.svc.CreateArc(b.Name, b.ObjectName)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, arc)
}

func (rt *Router) handleListArcs(w http.ResponseWriter, r *http.Request, _ []string) {
	arcs, err := rt.svc.ListArcs()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"arcs": arcs})
}

func (rt *Router) handleGetArc(w http.ResponseWriter, r *http.Request, c []string) {
	arc, err := rt.svc.GetArc(c[0])
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, 404, map[string]string{"error": "arc not found"})
		return
	}
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, arc)
}

func (rt *Router) handleAdvanceArc(w http.ResponseWriter, r *http.Request, c []string) {
	var b struct{ Status string `json:"status"` }
	if err := readJSON(r, &b); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	if err := rt.svc.AdvanceArc(c[0], b.Status); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (rt *Router) handleSealArc(w http.ResponseWriter, r *http.Request, c []string) {
	if err := rt.svc.SealArc(c[0]); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (rt *Router) handleImportObservation(w http.ResponseWriter, r *http.Request, c []string) {
	var b struct {
		StationID    string  `json:"station_id"`
		CatalogID    string  `json:"catalog_id"`
		ObsTimeUTC   string  `json:"obs_time_utc"`
		RA           float64 `json:"ra"`
		Dec          float64 `json:"dec"`
		RAErrArcsec  float64 `json:"ra_err_arcsec"`
		DecErrArcsec float64 `json:"dec_err_arcsec"`
	}
	if err := readJSON(r, &b); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	t, err := time.Parse(time.RFC3339, b.ObsTimeUTC)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "bad obs_time_utc: " + err.Error()})
		return
	}
	o, err := rt.svc.ImportObservation(c[0], b.StationID, b.CatalogID, t, b.RA, b.Dec, b.RAErrArcsec, b.DecErrArcsec)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, o)
}

func (rt *Router) handleListObservations(w http.ResponseWriter, r *http.Request, c []string) {
	obs, err := rt.svc.ListObservations(c[0])
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"observations": obs})
}

func (rt *Router) handleSetOrbit(w http.ResponseWriter, r *http.Request, c []string) {
	var b struct {
		A       float64 `json:"a"`
		E       float64 `json:"e"`
		IDeg    float64 `json:"i_deg"`
		OmDeg   float64 `json:"om_deg"`
		WDeg    float64 `json:"w_deg"`
		M0Deg   float64 `json:"m0_deg"`
		EpochJD float64 `json:"epoch_jd"`
		Source  string  `json:"source"`
	}
	if err := readJSON(r, &b); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	orb, err := rt.svc.SetOrbit(c[0], b.A, b.E, b.IDeg, b.OmDeg, b.WDeg, b.M0Deg, b.EpochJD, b.Source)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, orb)
}

func (rt *Router) handleGetOrbit(w http.ResponseWriter, r *http.Request, c []string) {
	orb, err := rt.svc.GetOrbit(c[0])
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, 404, map[string]string{"error": "orbit not found"})
		return
	}
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, orb)
}

func (rt *Router) handleListOrbits(w http.ResponseWriter, r *http.Request, c []string) {
	orbits, err := rt.svc.ListOrbits(c[0])
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"orbits": orbits})
}

func (rt *Router) handleComputeResiduals(w http.ResponseWriter, r *http.Request, c []string) {
	n, err := rt.svc.ComputeResiduals(c[0])
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"computed": n})
}

func (rt *Router) handleListResiduals(w http.ResponseWriter, r *http.Request, c []string) {
	res, err := rt.svc.GetResiduals(c[0])
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"residuals": res})
}

func (rt *Router) handleAnalyze(w http.ResponseWriter, r *http.Request, c []string) {
	a, err := rt.svc.Analyze(c[0])
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, a)
}

func (rt *Router) handleGetAttribution(w http.ResponseWriter, r *http.Request, c []string) {
	a, err := rt.svc.GetAttribution(c[0])
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, 404, map[string]string{"error": "attribution not found"})
		return
	}
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, a)
}

func (rt *Router) handlePublishSnapshot(w http.ResponseWriter, r *http.Request, c []string) {
	var b struct{ CatalogID string `json:"catalog_id"` }
	_ = readJSON(r, &b)
	snap, err := rt.svc.PublishOrbitSnapshot(c[0], b.CatalogID)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, snap)
}

func (rt *Router) handleListSnapshots(w http.ResponseWriter, r *http.Request, c []string) {
	snaps, err := rt.svc.ListSnapshots(c[0])
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"snapshots": snaps})
}

func (rt *Router) handleGetSnapshot(w http.ResponseWriter, r *http.Request, c []string) {
	snap, err := rt.svc.GetSnapshot(c[0])
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, 404, map[string]string{"error": "snapshot not found"})
		return
	}
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, snap)
}
