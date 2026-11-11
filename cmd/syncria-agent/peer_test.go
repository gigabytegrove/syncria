package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gigabytegrove/syncria/agentpeer"
)

func TestPeerInitAndReadToken(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "key")
	var out bytes.Buffer
	err := run([]string{"peer", "init", "-key-file", dest}, strings.NewReader(""), &out)
	if err != nil {
		t.Fatal(err)
	}
	secret, err := agentpeer.ReadKey(dest)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), secret) {
		t.Fatal("token not displayed once")
	}
	if err = run([]string{"peer", "init", "-key-file", dest}, strings.NewReader(""), &out); err == nil {
		t.Fatal("overwrote existing key")
	}
	if fi, err := os.Stat(dest); err != nil || fi.Mode().Perm() != 0600 {
		t.Fatalf("insecure key: %v %v", fi, err)
	}
}
func TestPeerServeFailsClosed(t *testing.T) {
	var out bytes.Buffer
	// Must not begin serving files without an enrolled device identity.
	err := run([]string{"peer", "serve", "-identity", filepath.Join(t.TempDir(), "missing.json")}, strings.NewReader(""), &out)
	if err == nil {
		t.Fatal("unenrolled agent allowed to serve")
	}
}
func TestPeerTokenInputRequiresFullLine(t *testing.T) {
	token := strings.Repeat("x", 43)
	value, err := readPeerToken(strings.NewReader(token + "\n"))
	if err != nil || value != token {
		t.Fatal(err)
	}
	if _, err := readPeerToken(strings.NewReader("short\n")); err == nil {
		t.Fatal("short token accepted")
	}
}
