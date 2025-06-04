package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

func validate_chirp(res http.ResponseWriter, req *http.Request) {
	type retSuccess struct {
		CleanedBody string `json:"cleaned_body"`
	}

	type retError struct {
		Err string `json:"error"`
	}

	type params struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(req.Body)
	parameters := params{}
	err := decoder.Decode(&parameters)
	res.Header().Set("Content-Type", "application/json")
	respError := retError{}
	if err != nil {
		respError.Err = "Something went wrong"
		res.WriteHeader(400)
		dat, _ := json.Marshal(respError)
		res.Write(dat)
		return
	}

	if len(parameters.Body) <= 140 {
		badWords := []string{"kerfuffle", "sharbert", "fornax"}
		for _, word := range badWords {
			parameters.Body = strings.ReplaceAll(parameters.Body, word, "****")
			parameters.Body = strings.ReplaceAll(parameters.Body, strings.ToUpper(word), "****")
			parameters.Body = strings.ReplaceAll(parameters.Body, strings.ToUpper(string(word[0]))+word[1:], "****")
		}
		resp := retSuccess{CleanedBody: parameters.Body}
		res.WriteHeader(200)
		dat, _ := json.Marshal(resp)
		res.Write(dat)
		return
	} else {
		respError.Err = "Chirp is too long"
		res.WriteHeader(400)
		dat, _ := json.Marshal(respError)
		res.Write(dat)
		return
	}

}

func main() {
	handler := http.ServeMux{}
	server := http.Server{Handler: &handler, Addr: ":8080"}
	handler.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	handler.HandleFunc("GET /api/healthz", health)
	handler.HandleFunc("GET /admin/metrics", getHits)
	handler.HandleFunc("POST /admin/reset", resetHits)
	handler.HandleFunc("POST /api/validate_chirp", validate_chirp)
	server.ListenAndServe()
}
