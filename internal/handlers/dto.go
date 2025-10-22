package handlers

type LongURLReq struct {
	LongURL string `json:"url"`
}

type ShortURLResp struct {
	ShortURL string `json:"result"`
}

type BathRequest struct {
	UUID    string `json:"correlation_id"`
	LongURL string `json:"original_url"`
}

type BathResponse struct {
	UUID     string `json:"correlation_id"`
	ShortURL string `json:"short_url"`
}
