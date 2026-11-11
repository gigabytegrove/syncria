package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAgentSyncCredentialIsolationAndDeletion(t *testing.T) {
	dir := t.TempDir()
	config := filepath.Join(dir, "state.json")
	id := strings.Repeat("a", 32)
	token := "YXh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHg"
	if err := saveAgentPeerCredential(config, id, token); err != nil {
		t.Fatal(err)
	}
	val, err := readAgentPeerCredential(config, id)
	if err != nil || val != token {
		t.Fatalf("credential not round tripped: %v", err)
	}
	path, _ := agentPeerTokenPath(config, id)
	stat, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if stat.Mode().Perm() != 0600 {
		t.Fatalf("credential mode %o", stat.Mode().Perm())
	}
	if err = saveAgentPeerCredential(config, id, token); err == nil {
		t.Fatal("credential unexpectedly replaced")
	}
	if err = removeAgentPeerCredential(config, id); err != nil {
		t.Fatal(err)
	}
	if _, err = readAgentPeerCredential(config, id); err == nil {
		t.Fatal("deleted credential still accessible")
	}
}
func TestAgentSyncURLRejectsUnsafeTransport(t *testing.T) {
	for _, v := range []string{"http://example.com", "ftp://example.com", "https://user:pass@example.com", "https://example.com/resource", "https://example.com/?bad=true", "https://example.com/#x"} {
		if err := validateAgentPeerURL(v); err == nil {
			t.Errorf("accepted %s", v)
		}
	}
	for _, v := range []string{"https://nas.example.com:9774", "http://localhost:9774", "http://127.0.0.1:9774"} {
		if err := validateAgentPeerURL(v); err != nil {
			t.Errorf("rejected %s: %v", v, err)
		}
	}
}
