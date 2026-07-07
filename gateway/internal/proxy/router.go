package proxy

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"gateway/internal/admission"
	"gateway/internal/auth"
	"gateway/internal/ratelimit"
)

type Route struct {
	Prefix     string
	Target     string
	Auth       bool
	Roles      []string
	WriteRoles []string
	Admission  bool
}

type Gateway struct {
	routes            []Route
	proxies           map[string]*httputil.ReverseProxy
	verifier          *auth.Verifier
	admission         *admission.Verifier
	admissionEnforced bool
	limiter           *ratelimit.Limiter
	log               *slog.Logger
}

func New(routes []Route, verifier *auth.Verifier, adm *admission.Verifier, admissionEnforced bool, limiter *ratelimit.Limiter, log *slog.Logger) (*Gateway, error) {
	proxies := make(map[string]*httputil.ReverseProxy)
	for _, r := range routes {
		if _, ok := proxies[r.Target]; ok {
			continue
		}
		u, err := url.Parse(r.Target)
		if err != nil {
			return nil, err
		}
		proxies[r.Target] = httputil.NewSingleHostReverseProxy(u)
	}
	return &Gateway{routes: routes, proxies: proxies, verifier: verifier, admission: adm, admissionEnforced: admissionEnforced, limiter: limiter, log: log}, nil
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/healthz" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
		return
	}

	if g.limiter != nil && !g.limiter.Allow(r.Context(), clientIP(r)) {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	route := g.match(r.URL.Path)
	if route == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	r.Header.Del("X-User-Id")
	r.Header.Del("X-User-Role")

	write := !isRead(r.Method)
	needAuth := route.Auth || len(route.Roles) > 0 || (write && len(route.WriteRoles) > 0)

	var claims *auth.Claims
	if needAuth {
		tok := bearer(r)
		if tok == "" {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		c, err := g.verifier.Verify(tok)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		claims = c
	}

	if len(route.Roles) > 0 && !contains(route.Roles, claims.Role) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if write && len(route.WriteRoles) > 0 && !contains(route.WriteRoles, claims.Role) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if route.Admission && g.admissionEnforced {
		at := r.Header.Get("X-Admission-Token")
		if at == "" {
			http.Error(w, "waiting room admission required", http.StatusForbidden)
			return
		}
		ac, err := g.admission.Verify(at)
		if err != nil || (claims != nil && ac.UserID != claims.UserID) {
			http.Error(w, "invalid admission token", http.StatusForbidden)
			return
		}
	}

	if claims != nil {
		r.Header.Set("X-User-Id", claims.UserID)
		r.Header.Set("X-User-Role", claims.Role)
	}

	g.proxies[route.Target].ServeHTTP(w, r)
}

func (g *Gateway) match(path string) *Route {
	var best *Route
	for i := range g.routes {
		if strings.HasPrefix(path, g.routes[i].Prefix) {
			if best == nil || len(g.routes[i].Prefix) > len(best.Prefix) {
				best = &g.routes[i]
			}
		}
	}
	return best
}

func isRead(method string) bool {
	return method == http.MethodGet || method == http.MethodHead
}

func bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if after, ok := strings.CutPrefix(h, "Bearer "); ok {
		return strings.TrimSpace(after)
	}
	return ""
}

func contains(set []string, v string) bool {
	for _, s := range set {
		if s == v {
			return true
		}
	}
	return false
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	host := r.RemoteAddr
	if i := strings.LastIndexByte(host, ':'); i >= 0 {
		return host[:i]
	}
	return host
}
