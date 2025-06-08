package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/maganor/Chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	queries        *database.Queries
	jwt_token      string
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
	if os.Getenv("PLATFORM") != "dev" {
		res.WriteHeader(403)
		return
	}
	apiCfg.resetHits()
	err := apiCfg.queries.DeleteUser(req.Context())
	if err != nil {
		fmt.Println(err)
		res.WriteHeader(400)
		res.Write([]byte("Something went wrong"))
		return
	}
	res.WriteHeader(200)
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	apiCfg.queries = database.New(db)
	apiCfg.jwt_token = os.Getenv("JWT_TOKEN")
	handler := http.ServeMux{}
	server := http.Server{Handler: &handler, Addr: ":8080"}
	handler.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	handler.HandleFunc("GET /api/healthz", health)
	handler.HandleFunc("GET /admin/metrics", getHits)
	handler.HandleFunc("POST /admin/reset", resetHits)
	handler.HandleFunc("POST /api/users", CreateUser)
	handler.HandleFunc("POST /api/revoke", RevokeToken)
	handler.HandleFunc("POST /api/refresh", RefreshUser)
	handler.HandleFunc("POST /api/login", Login)
	handler.HandleFunc("POST /api/chirps", CreateChirp)
	handler.HandleFunc("GET /api/chirps", GetChirps)
	handler.HandleFunc("GET /api/chirps/{id}", GetChirp)
	server.ListenAndServe()
}
