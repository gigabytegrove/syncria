package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNASVersionEndpointRequiresAdministrator(t *testing.T) {
	a := &App{}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/agent-sync/versions?job_id=whatever", nil)
	response := httptest.NewRecorder()
	a.agentVersionAPI(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected authorization before NAS lookup; got %d", response.Code)
	}
	if got := response.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("sensitive listing cacheable: %q", got)
	}
}
