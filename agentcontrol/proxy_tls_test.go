package agentcontrol

import (
	"crypto/tls"
	"net/http/httptest"
	"testing"
)

func TestProxyHeaderRequiresExplicitTrustedPeer(t *testing.T) {
	t.Setenv("SYNCRIA_TRUSTED_HTTPS_PROXY", "1")
	t.Setenv("SYNCRIA_TRUSTED_PROXY_CIDRS", "")
	req := httptest.NewRequest("POST", "http://syncria.example/api/v1/agents/heartbeat", nil)
	req.RemoteAddr = "198.51.100.99:32345"
	req.Header.Set("X-Forwarded-Proto", "https")
	if secureAgentRequest(req) {
		t.Fatal("untrusted public peer spoofed TLS")
	}
	req.RemoteAddr = "127.0.0.1:43210"
	if !secureAgentRequest(req) {
		t.Fatal("explicitly trusted loopback proxy denied")
	}
	req.RemoteAddr = "192.168.0.20:22365"
	if secureAgentRequest(req) {
		t.Fatal("unconfigured LAN peer spoofed TLS")
	}
	t.Setenv("SYNCRIA_TRUSTED_PROXY_CIDRS", "192.168.0.20/32")
	if !secureAgentRequest(req) {
		t.Fatal("explicit CIDR proxy denied")
	}
	req.RemoteAddr = "192.168.0.21:22365"
	if secureAgentRequest(req) {
		t.Fatal("peer outside explicit CIDR accepted")
	}
	req.RemoteAddr = "192.168.0.20:22365"
	req.Header.Set("X-Forwarded-Proto", "http")
	if secureAgentRequest(req) {
		t.Fatal("plaintext forwarded scheme accepted")
	}
	req.Header.Set("X-Forwarded-Proto", "https")
	t.Setenv("SYNCRIA_TRUSTED_HTTPS_PROXY", "0")
	if secureAgentRequest(req) {
		t.Fatal("disabled trusted proxy flag ignored")
	}
}
func TestDirectTLSAndLoopbackAreUnaffected(t *testing.T) {
	t.Setenv("SYNCRIA_TRUSTED_HTTPS_PROXY", "0")
	r := httptest.NewRequest("POST", "https://syncria.example/api/v1/agents/heartbeat", nil)
	r.TLS = &tls.ConnectionState{}
	r.RemoteAddr = "203.0.113.8:33000"
	if !secureAgentRequest(r) {
		t.Fatal("direct TLS denied")
	}
	local := httptest.NewRequest("POST", "http://localhost:9764/api/v1/agents/heartbeat", nil)
	local.RemoteAddr = "127.0.0.1:1234"
	if !secureAgentRequest(local) {
		t.Fatal("local developer connection denied")
	}
	local.Host = "syncria.example"
	if secureAgentRequest(local) {
		t.Fatal("spoofed external hostname accepted as local connection")
	}
}
