package handlers

import "github.com/google/uuid"

type LongURLReq struct {
	LongURL string `json:"url"`
}

type ShortURLResp struct {
	ShortURL string `json:"result"`
}

type BatchRequest struct {
	UUID    uuid.UUID `json:"correlation_id"`
	LongURL string    `json:"original_url"`
}

type BatchResponse struct {
	UUID     uuid.UUID `json:"correlation_id"`
	ShortURL string    `json:"short_url"`
}

type UserShortURLs struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"original_url"`
}
