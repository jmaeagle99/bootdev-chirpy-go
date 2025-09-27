package main

import (
	"io"
	"log"
	"net/http"
)

func main() {
	const contentRoot = "."
	const port = "8080"

	apiCfg := apiConfig{}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(contentRoot)))))
	mux.HandleFunc("GET /api/healthz", readinessHandler)
	mux.HandleFunc("GET /api/metrics", apiCfg.getHitsHandler)
	mux.HandleFunc("POST /api/reset", apiCfg.resetHitsHandler)

	server := http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, "OK")
}
