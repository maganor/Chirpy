package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

var apiCfg = apiConfig{}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) resetHits() {
	cfg.fileserverHits.Store(0)
}

func health(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(200)
	res.Write([]byte("OK"))
}

func getHits(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/html")
	res.WriteHeader(200)
	value := apiCfg.fileserverHits.Load()
	resp := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, value)
	res.Write([]byte(resp))
}

func resetHits(res http.ResponseWriter, req *http.Request) {
	apiCfg.resetHits()
	res.WriteHeader(200)
}

func main() {
	handler := http.ServeMux{}
	server := http.Server{Handler: &handler, Addr: ":8080"}
	handler.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	handler.HandleFunc("GET /api/healthz", health)
	handler.HandleFunc("GET /admin/metrics", getHits)
	handler.HandleFunc("POST /admin/reset", resetHits)
	server.ListenAndServe()
}
