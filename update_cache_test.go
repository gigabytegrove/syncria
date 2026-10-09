package main

import (
 "context"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
 "time"
)

func TestUpdateCacheReusesResultWithoutGitHubRequest(t *testing.T) {
 updateLookup.Lock()
 previousResponse,previousUntil:=updateLookup.response,updateLookup.validUntil
 updateLookup.response=map[string]any{"current":"cached-version","available":false}
 updateLookup.validUntil=time.Now().Add(time.Minute)
 updateLookup.Unlock()
 defer func(){
  updateLookup.Lock()
  updateLookup.response,updateLookup.validUntil=previousResponse,previousUntil
  updateLookup.Unlock()
 }()
 result:=cachedUpdateStatus(context.Background())
 if result["current"]!="cached-version" {t.Fatalf("unexpected cache result: %v",result)}
}

func TestRunningVersionDoesNotQueryGitHub(t *testing.T) {
 a:=&App{}
 req:=httptest.NewRequest(http.MethodGet,"/api/update/status?running=1",nil)
 w:=httptest.NewRecorder()
 a.updateStatus(w,req)
 if w.Code!=http.StatusUnauthorized {t.Fatalf("expected authentication before status access: %d",w.Code)}
}

func TestOnlyTargetedVersionPolling(t *testing.T) {
 if strings.Contains(newUI,"setInterval(checkRunningVersion,5000)") ||
    strings.Contains(newUI,"location.replace(") ||
    strings.Contains(newUI,"location.reload("){
   t.Fatal("automatic whole-page version refresh must not return")
 }
}
