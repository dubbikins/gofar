package routers

import (
	"net/http"

	"github.com/dubbikins/glam"
)

func HealthCheckRouter() *glam.Router {
	router := glam.NewRouter()
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		return
	})
	return router
}
