package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jmaeagle99/chirpy/internal/auth"
	"github.com/jmaeagle99/chirpy/internal/database"
)

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUserRequest struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

type UserResponse struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token,omitempty"`
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	request := CreateUserRequest{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&request)
	if err != nil {
		writeServerError(w)
		return
	}

	hashed_password, err := auth.HashPassword(request.Password)
	if err != nil {
		writeServerError(w)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          request.Email,
		HashedPassword: hashed_password,
	})
	if err != nil {
		writeServerError(w)
		return
	}

	writeAsJson(
		w,
		convertUser(user, ""),
		http.StatusCreated)
}

func (cfg *apiConfig) loginUser(w http.ResponseWriter, r *http.Request) {
	const MaxExpirationInSeconds = 3600 // 1 hour
	request := LoginUserRequest{}
	request.ExpiresInSeconds = MaxExpirationInSeconds // default

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&request)
	if err != nil {
		writeServerError(w)
		return
	}

	// Cap the expiration
	if request.ExpiresInSeconds > MaxExpirationInSeconds {
		request.ExpiresInSeconds = MaxExpirationInSeconds
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), request.Email)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	isMatch, err := auth.CheckPasswordHash(request.Password, user.HashedPassword)
	if err != nil || !isMatch {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := auth.MakeJWT(
		user.ID,
		cfg.tokenSecret,
		time.Duration(request.ExpiresInSeconds)*time.Second,
	)
	if err != nil || !isMatch {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	writeAsJson(
		w,
		convertUser(user, token),
		http.StatusOK)
}

func convertUser(user database.User, token string) UserResponse {
	response := UserResponse{
		Id:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	if len(token) > 0 {
		response.Token = token
	}
	return response
}
