package main

import (
	"net/http"
)

func health(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(200)
	res.Write([]byte("OK"))
}

func main() {
	handler := http.ServeMux{}
	server := http.Server{Handler: &handler, Addr: ":8080"}
	handler.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	handler.HandleFunc("/healthz", health)
	server.ListenAndServe()
}
