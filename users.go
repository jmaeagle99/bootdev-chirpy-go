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

type UpdateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	Id           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
}

type AccessTokenResponse struct {
	Token string `json:"token"`
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
		convertUser(user, "", ""),
		http.StatusCreated)
}

func (cfg *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
	userId, err := cfg.validateUserAccess(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	request := UpdateUserRequest{}

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&request)
	if err != nil {
		writeServerError(w)
		return
	}

	hashed_password, err := auth.HashPassword(request.Password)
	if err != nil {
		writeServerError(w)
		return
	}

	user, err := cfg.db.UpdateEmailAndPassword(r.Context(), database.UpdateEmailAndPasswordParams{
		ID:             userId,
		Email:          request.Email,
		HashedPassword: hashed_password,
	})
	if err != nil {
		writeServerError(w)
		return
	}

	writeAsJson(
		w,
		convertUser(user, "", ""),
		http.StatusOK)
}

func (cfg *apiConfig) loginUser(w http.ResponseWriter, r *http.Request) {
	request := LoginUserRequest{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&request)
	if err != nil {
		writeServerError(w)
		return
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

	refresh_token, err := auth.MakeRefreshToken()
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	_, err = cfg.db.RegisterRefreshToken(
		r.Context(),
		database.RegisterRefreshTokenParams{
			Token:     refresh_token,
			UserID:    user.ID,
			ExpiresAt: time.Now().UTC().Add(60 * 24 * time.Hour),
		})
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	access_token, err := auth.MakeJWT(
		user.ID,
		cfg.tokenSecret,
		time.Hour,
	)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	writeAsJson(
		w,
		convertUser(user, access_token, refresh_token),
		http.StatusOK)
}

func (cfg *apiConfig) getAccessToken(w http.ResponseWriter, r *http.Request) {
	refresh_token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	user, err := cfg.db.GetUserByRefreshToken(r.Context(), refresh_token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	access_token, err := auth.MakeJWT(
		user.ID,
		cfg.tokenSecret,
		time.Hour,
	)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	writeAsJson(
		w,
		AccessTokenResponse{
			Token: access_token,
		},
		http.StatusOK,
	)
}

func (cfg *apiConfig) revokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	refresh_token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), refresh_token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func convertUser(user database.User, access_token string, refresh_token string) UserResponse {
	return UserResponse{
		Id:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        access_token,
		RefreshToken: refresh_token,
	}
}

func (cfg *apiConfig) validateUserAccess(r *http.Request) (uuid.UUID, error) {
	access_token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		return uuid.Nil, err
	}

	return auth.ValidateJWT(access_token, cfg.tokenSecret)
}
