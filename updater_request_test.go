package main

import (
 "context"
 "io"
 "net/http"
 "net/http/httptest"
 "os"
 "strings"
 "testing"
)

func TestGitHubRequestRetriesAnonymouslyOnRejectedToken(t *testing.T) {
 t.Setenv("SYNCRIA_GITHUB_TOKEN","bad-token")
 calls:=0
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  calls++
  if r.Header.Get("Authorization")!=""{http.Error(w,"bad credentials",http.StatusForbidden);return}
  w.Header().Set("Content-Type","application/json")
  io.WriteString(w,`{"ok":true}`)
 }))
 defer server.Close()
 resp,err:=releaseRequest(context.Background(),server.URL,"application/vnd.github+json")
 if err!=nil{t.Fatal(err)}
 defer resp.Body.Close()
 if resp.StatusCode!=http.StatusOK||calls!=2{t.Fatalf("status %d, calls %d",resp.StatusCode,calls)}
}

func TestGitHubRequestDoesNotRetryAnonymousOnSuccess(t *testing.T) {
 t.Setenv("SYNCRIA_GITHUB_TOKEN","valid-token")
 calls:=0
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  calls++
  if r.Header.Get("Authorization")!="Bearer valid-token"{t.Error("missing authorization")}
  w.WriteHeader(http.StatusOK)
 }))
 defer server.Close()
 resp,err:=releaseRequest(context.Background(),server.URL,"application/vnd.github+json")
 if err!=nil{t.Fatal(err)}
 defer resp.Body.Close()
 if calls!=1 {t.Fatalf("unexpected request count %d",calls)}
}
func TestGitHub403ErrorIncludesRateLimit(t *testing.T) {
 r:=httptest.NewRecorder()
 r.Header().Set("X-RateLimit-Remaining","0")
 r.WriteHeader(http.StatusForbidden)
 _,_=io.WriteString(r,`{"message":"API rate limit exceeded"}`)
 err:=githubAPIError(r.Result(),"GitHub tag lookup")
 if !strings.Contains(err.Error(),"rate limit"){t.Fatal(err)}
}
func TestGitHub403ErrorReportsActualMessage(t *testing.T) {
 r:=httptest.NewRecorder()
 r.WriteHeader(http.StatusForbidden)
 _,_=io.WriteString(r,`{"message":"Resource not accessible by integration"}`)
 err:=githubAPIError(r.Result(),"GitHub tag lookup")
 if !strings.Contains(err.Error(),"Resource not accessible"){t.Fatal(err)}
}
func TestGitHubRequestWithoutToken(t *testing.T) {
 os.Unsetenv("SYNCRIA_GITHUB_TOKEN")
 t.Setenv("SYNCRIA_GITHUB_TOKEN","")
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if r.Header.Get("Authorization")!=""{t.Error("unexpected authorization")}
  w.WriteHeader(http.StatusOK)
 }))
 defer server.Close()
 resp,err:=releaseRequest(context.Background(),server.URL,"application/vnd.github+json")
 if err!=nil{t.Fatal(err)}
 resp.Body.Close()
}
