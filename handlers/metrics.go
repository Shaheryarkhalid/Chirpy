package handlers

import (
	"Chirpy/helpers"
	"Chirpy/internal/config"
	"fmt"
	"net/http"
)

type MetricsHandler struct {
	*config.ApiConfig
}

func (metricsHandler *MetricsHandler) MiddlewareMatricInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		metricsHandler.FileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})

}
func (metricsHandler *MetricsHandler) HandlerMetrics(respWriter http.ResponseWriter, req *http.Request) {
	respWriter.Header().Set("Content-Type", "text/html")
	respWriter.Header().Set("Cache-Control", "no-cache")
	respWriter.WriteHeader(200)
	respWriter.Write([]byte(fmt.Sprintf(
		`<html>
		  <body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		  </body>
		</html>`, metricsHandler.FileServerHits.Load())))
}
func (metricsHandler *MetricsHandler) HandlerReset(respWriter http.ResponseWriter, req *http.Request) {
	metricsHandler.FileServerHits.Store(0)
	if metricsHandler.ApiConfig.Platform != "dev" {
		helpers.RespondWithError(respWriter, 403, "403 Forbidden")
		return
	}
	err := metricsHandler.DB.DeleteAllUsers(req.Context())
	if err != nil {
		metricsHandler.Logger.Printf("Error trying to reset the users table:  %v\n", err)
		helpers.RespondWithError(respWriter, 500, "Some thing went wrong while trying to reset users db.")
		return
	}
	respWriter.Header().Set("Content-Type", "text/plain")
	respWriter.Header().Set("Cache-Control", "no-cache")
	respWriter.WriteHeader(200)
	respWriter.Write([]byte("Metrics Reset Successfull."))
}
