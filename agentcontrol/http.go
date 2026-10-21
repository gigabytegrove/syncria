package agentcontrol

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// Handler deliberately exposes no filesystem operation. Only approved,
// authenticated device identities can send heartbeat metadata.
type Handler struct {
	Store          *Store
	Admin          func(*http.Request) bool
	SameOriginPost func(*http.Request) bool
}

func reply(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func decode(w http.ResponseWriter, r *http.Request, out any) bool {
	if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		http.Error(w, "application/json required", http.StatusUnsupportedMediaType)
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(out); err != nil {
		http.Error(w, "invalid JSON", 400)
		return false
	}
	var extra any
	if err := d.Decode(&extra); !errors.Is(err, io.EOF) {
		http.Error(w, "extra JSON", 400)
		return false
	}
	return true
}
func secureAgentRequest(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
			// Avoid treating traffic forwarded from an HTTP reverse proxy as
			// a secure local agent connection solely because its TCP peer is local.
			requestHost := r.Host
			if h, _, e := net.SplitHostPort(requestHost); e == nil {
				requestHost = h
			}
			if strings.EqualFold(requestHost, "localhost") ||
				(net.ParseIP(requestHost) != nil && net.ParseIP(requestHost).IsLoopback()) {
				return true
			}
		}
	}
	// Do not trust a client-supplied X-Forwarded-Proto header on a direct
	// HTTP connection. The actual TCP peer MUST be an explicitly trusted
	// reverse proxy; a toggle alone must never authorize internet peers.
	if os.Getenv("SYNCRIA_TRUSTED_HTTPS_PROXY") != "1" ||
		!strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") || err != nil {
		return false
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	if ip.IsLoopback() {
		return true
	}
	for _, entry := range strings.Split(os.Getenv("SYNCRIA_TRUSTED_PROXY_CIDRS"), ",") {
		_, network, parseErr := net.ParseCIDR(strings.TrimSpace(entry))
		if parseErr == nil && network.Contains(ip) {
			return true
		}
	}
	return false
}
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if h.Store == nil {
		http.Error(w, "agent registry unavailable", 503)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/agents/")
	// Device endpoints use bearer credentials and MUST NOT accept browser sessions.
	if path == "enroll" || path == "heartbeat" {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			http.Error(w, "method not allowed", 405)
			return
		}
		if !secureAgentRequest(r) {
			http.Error(w, "TLS required for agent credentials", 403)
			return
		}
		if path == "enroll" {
			var x struct {
				Token    string `json:"token"`
				Platform string `json:"platform"`
				Arch     string `json:"arch"`
				Version  string `json:"version"`
			}
			if !decode(w, r, &x) {
				return
			}
			d, secret, err := h.Store.Redeem(x.Token, x.Platform, x.Arch, x.Version)
			if err != nil {
				status := 401
				if errors.Is(err, ErrPending) {
					status = 403
				}
				if errors.Is(err, ErrUsed) {
					status = 409
				}
				http.Error(w, err.Error(), status)
				return
			}
			reply(w, 201, map[string]any{"device": d, "credential": secret})
			return
		}
		id := r.Header.Get("X-Syncria-Agent-ID")
		authorization := r.Header.Get("Authorization")
		if !strings.HasPrefix(authorization, "Bearer ") || strings.Contains(authorization[7:], " ") {
			http.Error(w, "unauthorized", 401)
			return
		}
		var x struct {
			Platform string `json:"platform"`
			Arch     string `json:"arch"`
			Version  string `json:"version"`
		}
		if !decode(w, r, &x) {
			return
		}
		d, err := h.Store.Heartbeat(id, strings.TrimPrefix(authorization, "Bearer "), x.Platform, x.Arch, x.Version)
		if err != nil {
			http.Error(w, "unauthorized", 401)
			return
		}
		reply(w, 200, map[string]any{"device": d})
		return
	}
	if h.Admin == nil || !h.Admin(r) {
		http.Error(w, "unauthorized", 401)
		return
	}
	if r.Method == http.MethodPost && (h.SameOriginPost == nil || !h.SameOriginPost(r)) {
		http.Error(w, "same-origin POST required", 403)
		return
	}
	switch {
	case path == "devices" && r.Method == http.MethodGet:
		reply(w, 200, map[string]any{"devices": h.Store.ListDevices()})
	case path == "invitations" && r.Method == http.MethodGet:
		reply(w, 200, map[string]any{"invitations": h.Store.ListInvitations()})
	case path == "audit" && r.Method == http.MethodGet:
		reply(w, 200, map[string]any{"events": h.Store.Events()})
	case path == "invitations" && r.Method == http.MethodPost:
		var x struct {
			Name       string `json:"name"`
			TTLSeconds int    `json:"ttl_seconds"`
		}
		if !decode(w, r, &x) {
			return
		}
		if x.TTLSeconds == 0 {
			x.TTLSeconds = 600
		}
		inv, token, err := h.Store.Issue(x.Name, time.Duration(x.TTLSeconds)*time.Second)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		reply(w, 201, map[string]any{"invitation": inv, "token": token})
	case strings.HasPrefix(path, "invitations/") && strings.HasSuffix(path, "/approve") && r.Method == http.MethodPost:
		id := strings.TrimSuffix(strings.TrimPrefix(path, "invitations/"), "/approve")
		if id == "" || strings.Contains(id, "/") {
			http.Error(w, "not found", 404)
			return
		}
		if err := h.Store.Approve(id); err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		reply(w, 200, map[string]any{"ok": true})
	case strings.HasPrefix(path, "devices/") && strings.HasSuffix(path, "/revoke") && r.Method == http.MethodPost:
		id := strings.TrimSuffix(strings.TrimPrefix(path, "devices/"), "/revoke")
		if id == "" || strings.Contains(id, "/") {
			http.Error(w, "not found", 404)
			return
		}
		if err := h.Store.Revoke(id); err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		reply(w, 200, map[string]any{"ok": true})
	default:
		http.Error(w, "not found", 404)
	}
}
