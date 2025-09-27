package main

import (
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/jmaeagle99/chirpy/internal/database"
)

type apiConfig struct {
	dbQueries      *database.Queries
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) getHitsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	template := `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`
	fmt.Fprintf(w, template, cfg.fileserverHits.Load())
}

func (cfg *apiConfig) resetHitsHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
}
