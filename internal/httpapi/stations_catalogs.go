package httpapi

import (
	"net/http"
)

func (rt *Router) handleCreateStation(w http.ResponseWriter, r *http.Request, _ []string) {
	var b struct {
		Name      string  `json:"name"`
		Code      string  `json:"code"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Height    float64 `json:"height"`
	}
	if err := readJSON(r, &b); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	st, err := rt.svc.CreateStation(b.Name, b.Code, b.Latitude, b.Longitude, b.Height)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, st)
}

func (rt *Router) handleListStations(w http.ResponseWriter, r *http.Request, _ []string) {
	stations, err := rt.svc.ListStations()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"stations": stations})
}

func (rt *Router) handleCreateCatalog(w http.ResponseWriter, r *http.Request, _ []string) {
	var b struct {
		Name      string  `json:"name"`
		Epoch     string  `json:"epoch"`
		RefFrame  string  `json:"ref_frame"`
		BiasRA    float64 `json:"bias_ra_arcsec"`
		BiasDec   float64 `json:"bias_dec_arcsec"`
	}
	if err := readJSON(r, &b); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	c, err := rt.svc.CreateCatalog(b.Name, b.Epoch, b.RefFrame, b.BiasRA, b.BiasDec)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, c)
}

func (rt *Router) handleListCatalogs(w http.ResponseWriter, r *http.Request, _ []string) {
	cats, err := rt.svc.ListCatalogs()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"catalogs": cats})
}

func (rt *Router) handleCalibrate(w http.ResponseWriter, r *http.Request, c []string) {
	var b struct {
		RAZero  float64 `json:"ra_zero_arcsec"`
		DecZero float64 `json:"dec_zero_arcsec"`
		Note    string  `json:"note"`
	}
	if err := readJSON(r, &b); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	if err := rt.svc.ApplyCalibration(c[0], b.RAZero, b.DecZero, b.Note); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}
