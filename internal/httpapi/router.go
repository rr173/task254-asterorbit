// Package httpapi 提供 HTTP 接口层，路由统一以 /api 前缀，覆盖弧段、观测、台站、
// 星表、轨道、残差、归因、校准、快照与自检。
package httpapi

import (
	"encoding/json"
	"net/http"
	"regexp"

	"task254-asterorbit/internal/service"
)

type route struct {
	method  string
	pattern string
	re      *regexp.Regexp
	h       func(http.ResponseWriter, *http.Request, []string)
}

// Router 持有业务 Service 与路由表。
type Router struct {
	svc    *service.Service
	routes []route
}

// NewRouter 构造路由表（全部 /api 前缀）。
func NewRouter(svc *service.Service) *Router {
	rt := &Router{svc: svc}
	rt.routes = []route{
		{method: "POST", pattern: `^/api/arcs$`, re: nil, h: rt.handleCreateArc},
		{method: "GET", pattern: `^/api/arcs$`, re: nil, h: rt.handleListArcs},
		{method: "GET", pattern: `^/api/arcs/([^/]+)$`, re: nil, h: rt.handleGetArc},
		{method: "POST", pattern: `^/api/arcs/([^/]+)/advance$`, re: nil, h: rt.handleAdvanceArc},
		{method: "POST", pattern: `^/api/arcs/([^/]+)/seal$`, re: nil, h: rt.handleSealArc},
		{method: "POST", pattern: `^/api/stations$`, re: nil, h: rt.handleCreateStation},
		{method: "GET", pattern: `^/api/stations$`, re: nil, h: rt.handleListStations},
		{method: "POST", pattern: `^/api/catalogs$`, re: nil, h: rt.handleCreateCatalog},
		{method: "GET", pattern: `^/api/catalogs$`, re: nil, h: rt.handleListCatalogs},
		{method: "POST", pattern: `^/api/arcs/([^/]+)/observations$`, re: nil, h: rt.handleImportObservation},
		{method: "GET", pattern: `^/api/arcs/([^/]+)/observations$`, re: nil, h: rt.handleListObservations},
		{method: "POST", pattern: `^/api/arcs/([^/]+)/orbit$`, re: nil, h: rt.handleSetOrbit},
		{method: "GET", pattern: `^/api/arcs/([^/]+)/orbit$`, re: nil, h: rt.handleGetOrbit},
		{method: "GET", pattern: `^/api/arcs/([^/]+)/orbits$`, re: nil, h: rt.handleListOrbits},
		{method: "POST", pattern: `^/api/arcs/([^/]+)/compute$`, re: nil, h: rt.handleComputeResiduals},
		{method: "GET", pattern: `^/api/arcs/([^/]+)/residuals$`, re: nil, h: rt.handleListResiduals},
		{method: "POST", pattern: `^/api/arcs/([^/]+)/analyze$`, re: nil, h: rt.handleAnalyze},
		{method: "GET", pattern: `^/api/arcs/([^/]+)/attribution$`, re: nil, h: rt.handleGetAttribution},
		{method: "POST", pattern: `^/api/stations/([^/]+)/calibrate$`, re: nil, h: rt.handleCalibrate},
		{method: "POST", pattern: `^/api/arcs/([^/]+)/snapshot$`, re: nil, h: rt.handlePublishSnapshot},
		{method: "GET", pattern: `^/api/arcs/([^/]+)/snapshots$`, re: nil, h: rt.handleListSnapshots},
		{method: "GET", pattern: `^/api/snapshots/([^/]+)$`, re: nil, h: rt.handleGetSnapshot},
		{method: "GET", pattern: `^/api/selfcheck$`, re: nil, h: rt.handleSelfCheck},
	}
	for i := range rt.routes {
		rt.routes[i].re = regexp.MustCompile(rt.routes[i].pattern)
	}
	return rt
}

// ServeHTTP 按方法+路径分发。
func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range rt.routes {
		if route.method != r.Method {
			continue
		}
		m := route.re.FindStringSubmatch(r.URL.Path)
		if m == nil {
			continue
		}
		route.h(w, r, m[1:])
		return
	}
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found", "path": r.URL.Path})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func readJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}
