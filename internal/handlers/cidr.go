package handlers

import (
	"net"
	"net/http"
)

// cidrMiddleware is a middleware that check if user's requests is from trusted subnet (specified in configuration)
func (h Handler) cidrMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		getIP := r.Header.Get("X-Real-IP")
		if getIP == "" {
			h.zlog.Info().Msg("IP header is empty")
			http.Error(w, "invalid IP", http.StatusUnauthorized)
			return
		}
		ip := net.ParseIP(getIP)
		if ip == nil {
			h.zlog.Info().Str("IP", getIP).Msg("invalid IP")
			http.Error(w, "invalid IP", http.StatusUnauthorized)
			return
		}
		_, mask, err := net.ParseCIDR(h.cfg.GetTrustedSubnet())
		if err != nil {
			h.zlog.Error().Err(err).Str("IP", getIP).Msg("invalid IP")
			http.Error(w, "server error while checking", http.StatusInternalServerError)
			return
		}
		if !mask.Contains(ip) {
			h.zlog.Info().Msgf("IP %s is not trusted", getIP)
			http.Error(w, "IP is not trusted", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
