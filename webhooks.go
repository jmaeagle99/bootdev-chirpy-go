package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jmaeagle99/chirpy/internal/auth"
)

type WebhookEventData interface{}

type WebhookEventRequest struct {
	EventType string
	Data      WebhookEventData
}

type UserUpgradedEventData struct {
	UserId uuid.UUID `json:"user_id"`
}

func (request *WebhookEventRequest) UnmarshalJSON(data []byte) error {
	var eventMetadata struct {
		EventType string          `json:"event"`
		Data      json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(data, &eventMetadata); err != nil {
		return err
	}

	switch eventMetadata.EventType {
	case "user.upgraded":
		eventData := UserUpgradedEventData{}
		if err := json.Unmarshal(eventMetadata.Data, &eventData); err != nil {
			return err
		}
		request.Data = eventData
	}

	request.EventType = eventMetadata.EventType
	return nil
}

func (cfg *apiConfig) handleWebhook(w http.ResponseWriter, r *http.Request) {
	api_key, err := auth.GetAPIKey(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if api_key != cfg.polkaKey {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	request := WebhookEventRequest{}

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&request)
	if err != nil {
		writeServerError(w)
		return
	}

	switch eventData := request.Data.(type) {
	case UserUpgradedEventData:
		cfg.upgradeUserRed(w, r, eventData)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}
