package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/jmaeagle99/chirpy/internal/database"
)

type ChirpRequest struct {
	Body string `json:"body"`
	// Not secure, but will be fixed in future
	UserId uuid.UUID `json:"user_id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ChirpResponse struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

var bannedWords = []string{
	"kerfuffle",
	"sharbert",
	"fornax",
}

func Map[T any, U any](input []T, fn func(T) U) []U {
	result := make([]U, len(input))
	for i, v := range input {
		result[i] = fn(v)
	}
	return result
}

var bannedWordRegexps = Map(bannedWords, func(word string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(word) + `\b`)
})

func (cfg *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {
	request := ChirpRequest{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&request)
	if err != nil {
		writeServerError(w)
		return
	}

	if len(request.Body) > 140 {
		writeAsJson(
			w,
			ErrorResponse{
				Error: "Chirp is too long",
			},
			http.StatusBadRequest)
		return
	}

	content := request.Body
	for _, regexp := range bannedWordRegexps {
		content = regexp.ReplaceAllString(content, "****")
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   content,
		UserID: request.UserId,
	})
	if err != nil {
		writeServerError(w)
		return
	}

	writeAsJson(
		w,
		convertChirp(chirp),
		http.StatusCreated)
}

func (cfg *apiConfig) getAllChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		writeServerError(w)
		return
	}

	chirpsResponse := make([]ChirpResponse, len(chirps))
	for index, chirp := range chirps {
		chirpsResponse[index] = convertChirp(chirp)
	}

	writeAsJson(
		w,
		chirpsResponse,
		http.StatusOK)
}

func convertChirp(chirp database.Chirp) ChirpResponse {
	return ChirpResponse{
		Id:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}
}

func writeAsJson(w http.ResponseWriter, value interface{}, statucode int) {
	data, err := json.Marshal(value)
	if err != nil {
		writeServerError(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statucode)
	w.Write(data)
}

func writeServerError(w http.ResponseWriter) {
	errorResp := ErrorResponse{
		Error: "Something went wrong",
	}

	w.WriteHeader(http.StatusInternalServerError)

	data, err := json.Marshal(errorResp)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
