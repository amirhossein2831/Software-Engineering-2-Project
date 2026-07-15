package main

import (
	"net/http"
	"time"

	"gateway/internal/admission"
	"gateway/internal/auth"
	"gateway/internal/config"
	"gateway/internal/logger"
	"gateway/internal/proxy"
	"gateway/internal/ratelimit"
)

func main() {
	log := logger.New("gateway")

	authURL := config.Get("AUTH_URL", "http://localhost:8081")
	catalogURL := config.Get("CATALOG_URL", "http://localhost:8082")
	reservationURL := config.Get("RESERVATION_URL", "http://localhost:8083")
	checkoutURL := config.Get("CHECKOUT_URL", "http://localhost:8084")
	ticketingURL := config.Get("TICKETING_URL", "http://localhost:8085")
	realtimeURL := config.Get("REALTIME_URL", "http://localhost:8087")
	waitingURL := config.Get("WAITING_ROOM_URL", "http://localhost:8088")

	organizerRoles := []string{"organizer", "admin"}

	routes := []proxy.Route{
		{Prefix: "/auth", Target: authURL},
		{Prefix: "/admin", Target: authURL, Auth: true, Roles: []string{"admin"}},
		{Prefix: "/events", Target: catalogURL, WriteRoles: organizerRoles},
		{Prefix: "/venues", Target: catalogURL, WriteRoles: organizerRoles},
		{Prefix: "/seatmap", Target: reservationURL},
		{Prefix: "/holds", Target: reservationURL, Auth: true, Admission: true},
		{Prefix: "/checkout", Target: checkoutURL, Auth: true, Admission: true},
		{Prefix: "/orders", Target: checkoutURL, Auth: true},
		{Prefix: "/tickets", Target: ticketingURL, Auth: true},
		{Prefix: "/queue", Target: waitingURL, Auth: true},
		{Prefix: "/ws", Target: realtimeURL},
	}

	verifier := auth.NewVerifier(config.MustGet("JWT_SECRET"))
	adm := admission.NewVerifier(config.Get("ADMISSION_SECRET", "dev-admission-secret"))
	limiter := ratelimit.New(
		config.Get("REDIS_ADDR", "localhost:6379"),
		config.GetInt("RATE_LIMIT", 120),
		time.Duration(config.GetInt("RATE_WINDOW_SECONDS", 60))*time.Second,
	)

	admissionEnforced := config.Get("ADMISSION_ENFORCED", "false") == "true"

	gw, err := proxy.New(routes, verifier, adm, admissionEnforced, limiter, log)
	if err != nil {
		log.Error("gateway init failed", "err", err)
		panic(err)
	}

	handler := withCORS(gw, config.Get("ALLOWED_ORIGIN", "*"))

	addr := ":" + config.Get("GATEWAY_PORT", "8080")
	log.Info("gateway listening", "addr", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Error("server stopped", "err", err)
		panic(err)
	}
}

func withCORS(next http.Handler, origin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", origin)
		h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Admission-Token, Idempotency-Key")
		h.Set("Access-Control-Max-Age", "600")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
