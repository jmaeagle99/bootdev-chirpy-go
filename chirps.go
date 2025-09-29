package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jmaeagle99/chirpy/internal/database"
)

type ChirpRequest struct {
	Body string `json:"body"`
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
	userId, err := cfg.validateUserAccess(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	request := ChirpRequest{}

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&request)
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
		UserID: userId,
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

func (cfg *apiConfig) deleteChirp(w http.ResponseWriter, r *http.Request) {
	userId, err := cfg.validateUserAccess(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	chirpId, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		writeAsJson(
			w,
			ErrorResponse{
				Error: "chirpId is not valid",
			},
			http.StatusBadRequest)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if chirp.UserID != userId {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), chirpId)
	if err != nil {
		writeServerError(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) getAllChirps(w http.ResponseWriter, r *http.Request) {
	author_id_qparam := r.URL.Query().Get("author_id")

	var chirps []database.Chirp
	if len(author_id_qparam) > 0 {
		user_id, err := uuid.Parse(author_id_qparam)
		if err != nil {
			writeServerError(w)
			return
		}

		result, err := cfg.db.GetAllChirpsByUser(r.Context(), user_id)
		if err != nil {
			writeServerError(w)
			return
		}
		chirps = result
	} else {
		result, err := cfg.db.GetAllChirps(r.Context())
		if err != nil {
			writeServerError(w)
			return
		}
		chirps = result
	}

	sort_qparam := r.URL.Query().Get("sort")
	if len(sort_qparam) > 0 {
		switch sort_qparam {
		case "asc":
			sort.Slice(chirps, func(a, b int) bool {
				return chirps[a].CreatedAt.Before(chirps[b].CreatedAt)
			})
		case "desc":
			sort.Slice(chirps, func(a, b int) bool {
				return chirps[b].CreatedAt.Before(chirps[a].CreatedAt)
			})
		}
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

func (cfg *apiConfig) getChirp(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		writeAsJson(
			w,
			ErrorResponse{
				Error: "chirpId is not valid",
			},
			http.StatusBadRequest)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	writeAsJson(
		w,
		convertChirp(chirp),
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
