package handlers

import (
	"github.com/google/uuid"
)

// LongURLReq used to unmarshal provided json
type LongURLReq struct {
	LongURL string `json:"url"`
}

// ShortURLResp used to marshal the response with slug
type ShortURLResp struct {
	ShortURL string `json:"result"`
}

// BatchRequest used to unmarshal batch of multiple urls from json
type BatchRequest struct {
	UUID    uuid.UUID `json:"correlation_id"`
	LongURL string    `json:"original_url"`
}

// BatchResponse used to marshall slugs for response to batch urls request
type BatchResponse struct {
	UUID     uuid.UUID `json:"correlation_id"`
	ShortURL string    `json:"short_url"`
}

// UserShortURLs used for responding to requests where the requestor is the owner
type UserShortURLs struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"original_url"`
}

// Stats used for responding to number of slugs and users requests
type Stats struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}
