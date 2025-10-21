package model

type URL struct {
	ID          int    `json:"uuid" db:"id"`
	ShortURL    string `json:"short_url" db:"short_url"`
	OriginalURL string `json:"original_url" db:"original_url"`
}
