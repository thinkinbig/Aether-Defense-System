package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// HealthHandler handles GET /health requests for Kubernetes health checks.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	httpx.OkJsonCtx(r.Context(), w, map[string]string{
		"status": "ok",
	})
}
