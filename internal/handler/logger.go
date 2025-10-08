package handler

import (
	"net/http"
	"time"
)

// loggingResponseWriter structure for logging size and status code of responses
type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// Write is redeclared method for embedding size into response logging
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader is redeclared method for embedding status code into response logging
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// requestLogger is middleware logger using zerolog
func (h Handler) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // original http.ResponseWriter embedding
			responseData:   responseData,
		}

		next.ServeHTTP(&lw, r)

		h.zlog.
			Info().
			Str("method", r.Method).
			Str("url", r.RequestURI).
			Str("user_agent", r.UserAgent()).
			Int("size", responseData.size).
			Dur("elapsed_ms", time.Since(start)).
			Int("status", responseData.status).
			Msg("incoming request")
	})
}
