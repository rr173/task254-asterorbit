package httpapi

import "net/http"

func (rt *Router) handleSelfCheck(w http.ResponseWriter, r *http.Request, _ []string) {
	res := rt.svc.SelfCheck()
	code := 200
	if !res.OK {
		code = 503
	}
	writeJSON(w, code, res)
}
