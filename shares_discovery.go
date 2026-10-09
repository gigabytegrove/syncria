package main

import (
 "context"
 "encoding/json"
 "errors"
 "net"
 "net/http"
 "os/exec"
 "sort"
 "strings"
 "time"
)

// discoverSMBShares lists advertised disk shares without attempting to mount them.
// Discovery runs on the Syncria host, not inside the user's browser.
func discoverSMBShares(ctx context.Context, server string) ([]string, error) {
 ip := net.ParseIP(strings.TrimSpace(server))
 if ip == nil || !(ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast()) {
  return nil, errors.New("Enter a private/local IPv4 or IPv6 address")
 }
 if _, err := exec.LookPath("smbclient"); err != nil {
  return nil, errors.New("SMB discovery requires smbclient on the Syncria host")
 }
 ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
 defer cancel()
 // -N uses anonymous/guest authentication without prompting for a password.
 // SMB servers may decline anonymous enumeration even though mounting works.
 output, err := exec.CommandContext(ctx, "smbclient", "-L", "//"+ip.String(), "-g", "-N").CombinedOutput()
 if ctx.Err() != nil { return nil, errors.New("SMB discovery timed out") }
 if err != nil { return nil, errors.New("The server did not permit anonymous share discovery; enter the share name manually") }
 found := map[string]bool{}
 for _, line := range strings.Split(string(output), "\n") {
  parts := strings.Split(strings.TrimSpace(line), "|")
  if len(parts) < 2 || !strings.EqualFold(strings.TrimSpace(parts[0]), "Disk") { continue }
  name := strings.TrimSpace(parts[1])
  if shareName.MatchString(name) && !strings.HasSuffix(name, "$") { found[name]=true }
 }
 shares := make([]string,0,len(found))
 for name:=range found { shares=append(shares,name) }
 sort.Strings(shares)
 return shares,nil
}

func (a *App) shareDiscover(w http.ResponseWriter, r *http.Request) {
 w.Header().Set("Content-Type","application/json; charset=utf-8")
 w.Header().Set("Cache-Control","no-store")
 if r.Method!=http.MethodGet { http.Error(w,`{"error":"GET required"}`,http.StatusMethodNotAllowed);return }
 if !a.authenticated(r) { w.WriteHeader(http.StatusUnauthorized);_ = json.NewEncoder(w).Encode(map[string]string{"error":"Unauthorized"});return }
 shares,err:=discoverSMBShares(r.Context(),r.URL.Query().Get("server"))
 if err!=nil {
  w.WriteHeader(http.StatusBadRequest)
  _=json.NewEncoder(w).Encode(map[string]string{"error":err.Error()})
  return
 }
 _=json.NewEncoder(w).Encode(struct{Shares []string `json:"shares"`}{shares})
}
