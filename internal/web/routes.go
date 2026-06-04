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
		ip := h.clientIP(r)
		limiter := h.ipLimiter.getLimiter(ip)
		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func (h *Handler) clientIP(r *http.Request) string {
	if h.trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			if first := strings.TrimSpace(strings.Split(xff, ",")[0]); first != "" {
				return first
			}
		}
		if real := strings.TrimSpace(r.Header.Get("X-Real-IP")); real != "" {
			return real
		}
	}
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
	fs := http.FileServer(http.Dir("ui/static"))
	mux.HandleFunc("/privacy", h.privacyHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("/secret", h.rateLimit(h.createHandler))
	mux.HandleFunc("/confirm/", h.rateLimit(h.confirmHandler))
	mux.HandleFunc("/status/", h.rateLimit(h.statusHandler))
	mux.HandleFunc("/sse", h.rateLimit(h.SSEHandler))
	mux.HandleFunc("/", h.formHandler)

	return mux
}
