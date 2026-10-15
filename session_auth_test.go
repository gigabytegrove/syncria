package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestSessionDurationsAndCookieSecurity(t *testing.T) {
	a := &App{cfg: Config{Secret: b64rand()}}
	for _, tc := range []struct{name string; remember bool; lifetime time.Duration; persistent bool}{
		{"browser session", false, defaultSessionLifetime, false},
		{"remembered", true, rememberedSessionLifetime, true},
	} {
		t.Run(tc.name, func(t *testing.T){
			req := httptest.NewRequest(http.MethodGet, "http://example.test/", nil)
			req.Header.Set("X-Forwarded-Proto", "https")
			w := httptest.NewRecorder()
			a.setSession(w, req, tc.remember)
			cookies := w.Result().Cookies()
			if len(cookies) != 1 {t.Fatalf("got %d cookies", len(cookies))}
			c := cookies[0]
			if c.Name != "gdsync_session" || !c.HttpOnly || !c.Secure || c.SameSite != http.SameSiteLaxMode {
				t.Fatalf("unsafe cookie flags: %+v", c)
			}
			if tc.persistent {
				if c.MaxAge != int(tc.lifetime/time.Second) {t.Fatalf("remembered MaxAge: %d", c.MaxAge)}
			} else if c.MaxAge != 0 || !c.Expires.IsZero() {
				t.Fatalf("browser session became persistent: %+v", c)
			}
			part := strings.Split(c.Value, ".")
			if len(part) != 2 {t.Fatal("bad signature format")}
			raw,err := base64.RawURLEncoding.DecodeString(part[0])
			if err != nil {t.Fatal(err)}
			var data struct{Expires int64 `json:"expires"`; Nonce string `json:"nonce"`}
			if err := json.Unmarshal(raw,&data);err != nil {t.Fatal(err)}
			remaining := time.Until(time.Unix(data.Expires,0))
			if remaining < tc.lifetime-time.Minute || remaining > tc.lifetime+time.Minute || data.Nonce == "" {t.Fatalf("wrong signed lifetime: %v", remaining)}
			authReq := httptest.NewRequest(http.MethodGet,"https://example.test/",nil)
			authReq.AddCookie(c)
			if !a.authenticated(authReq) {t.Fatal("signed session rejected")}
			badCookie := *c
			badCookie.Value += "bad"
			badReq := httptest.NewRequest(http.MethodGet,"https://example.test/",nil)
			badReq.AddCookie(&badCookie)
			if a.authenticated(badReq) {t.Fatal("modified session accepted")}
		})
	}
}

func TestLoginRememberMeAndLogout(t *testing.T) {
	pwd := "correct horse battery staple"
	a := &App{cfg: Config{Secret: b64rand(),Salt:newID()}}
	a.cfg.PasswordHash = hashPass(pwd,a.cfg.Salt)
	for _, persistent := range []bool{false,true} {
		v := url.Values{"action":{"login"},"password":{pwd}}
		if persistent {v.Set("remember","on")}
		req := httptest.NewRequest("POST","https://example.test/action",strings.NewReader(v.Encode()))
		req.Header.Set("Content-Type","application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		a.action(w,req)
		if w.Code != http.StatusSeeOther {t.Fatalf("login failed: %d",w.Code)}
		cookie := w.Result().Cookies()[0]
		if (cookie.MaxAge > 0) != persistent {t.Fatalf("unexpected persistence for %v",persistent)}
		if !cookie.Secure {t.Fatal("TLS login did not set Secure")}
		logout := httptest.NewRequest("POST","https://example.test/action",strings.NewReader("action=logout"))
		logout.Header.Set("Content-Type","application/x-www-form-urlencoded")
		logout.AddCookie(cookie)
		out := httptest.NewRecorder()
		a.action(out,logout)
		if out.Code != http.StatusSeeOther {t.Fatalf("logout: %d",out.Code)}
		removed := out.Result().Cookies()[0]
		if removed.MaxAge >= 0 || !removed.Secure {t.Fatalf("cookie not cleared securely: %+v",removed)}
	}
}

func TestLoggedOutScreenContainsNoLiveStatus(t *testing.T) {
	a := &App{cfg: Config{PasswordHash:"configured",Secret:b64rand()}}
	req := httptest.NewRequest("GET","https://example.test/?msg=Sync%20started%20(check%20status)",nil)
	w := httptest.NewRecorder()
	a.home(w,req)
	if w.Code != http.StatusOK {t.Fatalf("home: %d",w.Code)}
	body := w.Body.String()
	for _, forbidden := range []string{"id=\"liveSystemBar\"","id=\"liveActivityFeed\"","/api/dashboard/status","/api/live","Sync started (check status)","<script>"} {
		if strings.Contains(body,forbidden) {t.Errorf("unauthenticated page exposes %q",forbidden)}
	}
	for _, needed := range []string{"name=\"remember\"","Sign in to Syncria","class=\"panel auth-card\""} {
		if !strings.Contains(body,needed) {t.Errorf("missing sign-in element %q",needed)}
	}
}

func TestLoginFormRetainsSameOriginProtection(t *testing.T) {
	a := &App{cfg: Config{PasswordHash:"configured",Secret:b64rand()}}
	req := httptest.NewRequest("POST","https://example.test/action",strings.NewReader("action=login&password=test"))
	req.Header.Set("Origin","https://attacker.invalid")
	req.Header.Set("Content-Type","application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	a.action(w,req)
	if w.Code != http.StatusMethodNotAllowed {t.Fatalf("cross-origin login accepted: %d",w.Code)}
}
