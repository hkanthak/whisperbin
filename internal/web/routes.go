package web

import (
	"net"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
	rate     rate.Limit
	burst    int
}

func newIPLimiter(r rate.Limit, b int) *ipLimiter {
	return &ipLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

func (l *ipLimiter) getLimiter(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, exists := l.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(l.rate, l.burst)
		l.limiters[ip] = limiter
	}
	return limiter
}

func (h *Handler) rateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)
		limiter := h.ipLimiter.getLimiter(ip)
		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func extractIP(r *http.Request) string {
	ip := r.RemoteAddr
	if ip == "" {
		return ""
	}
	if strings.Contains(ip, ":") {
		ip, _, _ = net.SplitHostPort(ip)
	}
	return ip
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir(projectRootPath("ui/static")))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("/secret", h.rateLimit(h.createHandler))
	mux.HandleFunc("/confirm/", h.rateLimit(h.confirmHandler))
	mux.HandleFunc("/status/", h.rateLimit(h.statusHandler))
	mux.HandleFunc("/ws", h.rateLimit(h.WebSocketHandler))
	mux.HandleFunc("/", h.formHandler)

	return mux
}
