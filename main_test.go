package main

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPassword(t *testing.T) {
	c := Config{Salt: newID()}
	c.PasswordHash = hashPass("correct horse battery staple", c.Salt)
	if !passwordOK("correct horse battery staple", c) || passwordOK("incorrect", c) {
		t.Fatal("password verification")
	}
}
func TestLocalTree(t *testing.T) {
	dir := t.TempDir()
	if e := os.Mkdir(filepath.Join(dir, "sub"), 0700); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(filepath.Join(dir, "sub", "a.txt"), []byte("hello"), 0600); e != nil {
		t.Fatal(e)
	}
	tree, e := localTree(dir)
	if e != nil {
		t.Fatal(e)
	}
	if !tree["sub"].Folder || tree["sub/a.txt"].LocalMD5 != "5d41402abc4b2a76b9719d911017c592" {
		t.Fatalf("unexpected tree: %v", tree)
	}
}
func TestSafeName(t *testing.T) {
	for _, s := range []string{"..", "../x", "a/b", "a\\b", "", ".sync-agent-trash"} {
		if safeName(s) {
			t.Errorf("unsafe name accepted: %q", s)
		}
	}
	if !safeName("hello.txt") {
		t.Fatal("valid name rejected")
	}
}
func TestAuthFlow(t *testing.T) {
	dir := t.TempDir()
	a := &App{file: filepath.Join(dir, "state.json"), cfg: Config{Secret: b64rand()}}
	req := httptest.NewRequest("POST", "http://localhost:8787/action", strings.NewReader("action=setup&password=longpassword123"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	a.action(w, req)
	if w.Code != 303 {
		t.Fatalf("setup: %d: %s", w.Code, w.Body.String())
	}
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("missing cookie")
	}
	req2 := httptest.NewRequest("GET", "http://localhost:8787/", nil)
	req2.AddCookie(cookies[0])
	if !a.authenticated(req2) {
		t.Fatal("session rejected")
	}
	req3 := httptest.NewRequest("POST", "http://localhost:8787/action", strings.NewReader("action=logout"))
	req3.Header.Set("Origin", "https://evil.example")
	w3 := httptest.NewRecorder()
	a.action(w3, req3)
	if w3.Code != 405 {
		t.Fatalf("CSRF POST accepted: %d", w3.Code)
	}
}
