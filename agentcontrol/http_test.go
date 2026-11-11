package agentcontrol

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func request(t *testing.T, h http.Handler, method, path, payload string, admin bool, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(payload))
	r.RemoteAddr = "192.0.2.22:1234"
	if payload != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	if admin {
		r.Header.Set("X-Test-Admin", "yes")
	}
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func TestAuthenticatedHTTPEnrollmentFlow(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "agents.json"))
	if err != nil {
		t.Fatal(err)
	}
	h := Handler{Store: store, Admin: func(r *http.Request) bool { return r.Header.Get("X-Test-Admin") == "yes" }, SameOriginPost: func(r *http.Request) bool { return r.Header.Get("Origin") != "https://evil.example" }}
	if w := request(t, h, "GET", "/api/v1/agents/devices", "", false, nil); w.Code != 401 {
		t.Fatal("unauthenticated admin list exposed")
	}
	w := request(t, h, "POST", "/api/v1/agents/invitations", `{"name":"Desk","ttl_seconds":600}`, true, map[string]string{"Origin": "https://evil.example"})
	if w.Code != 403 {
		t.Fatalf("cross origin admin POST allowed %d", w.Code)
	}
	w = request(t, h, "POST", "/api/v1/agents/invitations", `{"name":"Desk","ttl_seconds":600}`, true, nil)
	if w.Code != 201 {
		t.Fatalf("create %d: %s", w.Code, w.Body.String())
	}
	var issued struct {
		Invitation InvitationView `json:"invitation"`
		Token      string         `json:"token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &issued); err != nil {
		t.Fatal(err)
	}
	agentBody, _ := json.Marshal(map[string]string{"token": issued.Token, "platform": "linux", "arch": "amd64", "version": "1.0.0"})
	w = request(t, h, "POST", "/api/v1/agents/enroll", string(agentBody), false, nil)
	if w.Code != 403 {
		t.Fatalf("cleartext bearer permitted: %d", w.Code)
	}
	// TLS is provided by httptest's request object, as it would be by net/http.
	secureRequest := func(method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		r.TLS = &tls.ConnectionState{}
		if body != "" {
			r.Header.Set("Content-Type", "application/json")
		}
		for k, v := range headers {
			r.Header.Set(k, v)
		}
		wr := httptest.NewRecorder()
		h.ServeHTTP(wr, r)
		return wr
	}
	w = secureRequest("POST", "/api/v1/agents/enroll", string(agentBody), nil)
	if w.Code != 403 {
		t.Fatalf("unapproved token accepted: %d", w.Code)
	}
	w = request(t, h, "POST", "/api/v1/agents/invitations/"+issued.Invitation.ID+"/approve", "", true, nil)
	if w.Code != 200 {
		t.Fatalf("approve %d: %s", w.Code, w.Body.String())
	}
	w = secureRequest("POST", "/api/v1/agents/enroll", string(agentBody), nil)
	if w.Code != 201 {
		t.Fatalf("enroll %d: %s", w.Code, w.Body.String())
	}
	var result struct {
		Device     DeviceView `json:"device"`
		Credential string     `json:"credential"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	w = secureRequest("POST", "/api/v1/agents/heartbeat", `{"platform":"linux","arch":"amd64","version":"1.0.0"}`, map[string]string{"Authorization": "Bearer " + result.Credential, "X-Syncria-Agent-ID": result.Device.ID})
	if w.Code != 200 {
		t.Fatalf("heartbeat %d: %s", w.Code, w.Body.String())
	}
	w = request(t, h, "POST", "/api/v1/agents/devices/"+result.Device.ID+"/revoke", "", true, nil)
	if w.Code != 200 {
		t.Fatalf("revoke %d: %s", w.Code, w.Body.String())
	}
	w = secureRequest("POST", "/api/v1/agents/heartbeat", `{"platform":"linux","arch":"amd64","version":"1.0.0"}`, map[string]string{"Authorization": "Bearer " + result.Credential, "X-Syncria-Agent-ID": result.Device.ID})
	if w.Code != 401 {
		t.Fatalf("revoked heartbeat accepted %d", w.Code)
	}
	if w = request(t, h, "GET", "/api/v1/agents/audit", "", true, nil); w.Code != 200 {
		t.Fatal("audit unavailable")
	}
}
