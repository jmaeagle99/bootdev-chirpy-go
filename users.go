package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type CreateUserRequest struct {
	Email string `json:"email"`
}

type CreateUserResponse struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	request := CreateUserRequest{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&request)
	if err != nil {
		writeServerError(w)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), request.Email)
	if err != nil {
		writeServerError(w)
		return
	}

	writeAsJson(
		w,
		CreateUserResponse{
			Id:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
		http.StatusCreated)
}
