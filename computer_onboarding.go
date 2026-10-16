package main

import (
 "context"
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "errors"
 "fmt"
 "net"
 "net/http"
 "net/url"
 "strings"
 "time"
)

// Enrollment invitation exists only in hub memory and expires after 20 minutes.
// Once the agent passes credential + pinned-TLS verification, it is consumed.
type computerInvite struct {
 Label string
 Expires time.Time
 State string
 ComputerID string
 Message string
}
func (a *App) getComputerInvite(code string)(computerInvite,bool){
 a.computerInviteMu.Lock()
 defer a.computerInviteMu.Unlock()
 invite,ok:=a.computerInvites[code]
 if !ok||time.Now().After(invite.Expires){return computerInvite{},false}
 return *invite,true
}
func (a *App) computerInviteAPI(w http.ResponseWriter,r *http.Request){
 if !a.authenticated(r){computerError(w,401,"Sign in to Syncria");return}
 if r.Method=="GET"{
  inv,ok:=a.getComputerInvite(r.URL.Query().Get("code"))
  if !ok{computerError(w,404,"Invitation expired or not found");return}
  computerJSON(w,200,map[string]any{"status":inv.State,"label":inv.Label,"computer_id":inv.ComputerID,"message":inv.Message,"expires_at":inv.Expires});return
 }
 if r.Method!=http.MethodPost||!localPost(r){computerError(w,405,"Authenticated same-origin POST required");return}
 var req struct{ Label string `json:"label"`}
 if err:=computerDecode(r,&req);err!=nil{computerError(w,400,err.Error());return}
 req.Label=strings.TrimSpace(req.Label)
 if req.Label==""||len(req.Label)>80{computerError(w,400,"Enter a computer name (up to 80 characters)");return}
 code:=newID()+newID() // 256-bit unpredictable, one-time enrollment authority
 invite:=&computerInvite{Label:req.Label,Expires:time.Now().Add(20*time.Minute),State:"waiting"}
 a.computerInviteMu.Lock()
 if a.computerInvites==nil{a.computerInvites=map[string]*computerInvite{}}
 for k,v:=range a.computerInvites{if time.Now().After(v.Expires){delete(a.computerInvites,k)}}
 if len(a.computerInvites)>=64{a.computerInviteMu.Unlock();computerError(w,429,"Too many pending computer invitations");return}
 a.computerInvites[code]=invite
 a.computerInviteMu.Unlock()
 computerJSON(w,201,map[string]any{"code":code,"expires_at":invite.Expires,"label":invite.Label})
}
func safeComputerAddress(ipString string)bool{
 ip:=net.ParseIP(ipString)
 return ip!=nil && (ip.IsPrivate()||ip.IsLoopback()||ip.IsLinkLocalUnicast()) && !ip.IsUnspecified() && !ip.IsMulticast()
}
func (a *App) computerEnrollAPI(w http.ResponseWriter,r *http.Request){
 if r.Method!=http.MethodPost{computerError(w,405,"POST required");return}
 var req struct{
  Code string `json:"code"`
  Key string `json:"key"`
  Fingerprint string `json:"fingerprint"`
  Addresses []string `json:"addresses"`
  Port int `json:"port"`
 }
 if err:=computerDecode(r,&req);err!=nil{computerError(w,400,err.Error());return}
 req.Fingerprint=strings.ToLower(strings.ReplaceAll(strings.TrimSpace(req.Fingerprint),":",""))
 if len(req.Key)<32||len(req.Key)>256{computerError(w,400,"Invalid agent pairing key");return}
 if _,err:=pinnedComputerClient(req.Fingerprint);err!=nil{computerError(w,400,err.Error());return}
 if req.Port<1024||req.Port>65535||len(req.Addresses)==0||len(req.Addresses)>12{computerError(w,400,"Invalid agent connection details");return}
 a.computerInviteMu.Lock()
 inv,ok:=a.computerInvites[req.Code]
 if !ok||time.Now().After(inv.Expires){a.computerInviteMu.Unlock();computerError(w,403,"Enrollment invitation expired");return}
 if inv.State=="connected"{a.computerInviteMu.Unlock();computerError(w,409,"Invitation already used");return}
 if inv.State=="verifying"{a.computerInviteMu.Unlock();computerError(w,409,"Enrollment already in progress");return}
 inv.State="verifying"
 label:=inv.Label
 a.computerInviteMu.Unlock()
 success:=false
 defer func(){
  if success{return}
  a.computerInviteMu.Lock()
  if inv.State=="verifying"{inv.State="waiting"}
  a.computerInviteMu.Unlock()
 }()
 ctx,cancel:=context.WithTimeout(r.Context(),22*time.Second);defer cancel()
 var rec ComputerRecord
 var status struct{Roots []agentRootInfo `json:"roots"`}
 var last error=errors.New("no candidate LAN address could be reached")
 for _,ip:=range req.Addresses{
  if !safeComputerAddress(ip){continue}
  host:=net.JoinHostPort(ip,fmt.Sprintf("%d",req.Port))
  candidate:=ComputerRecord{ID:newID(),Label:label,URL:"https://"+host,Key:req.Key,Fingerprint:req.Fingerprint}
  if _,err:=validAgentURL(candidate.URL);err!=nil{continue}
  if err:=agentGET(ctx,candidate,"/v1/status",&status);err!=nil{last=err;continue}
  if len(status.Roots)==0{last=errors.New("agent has no explicitly shared folders");continue}
  rec=candidate;break
 }
 if rec.ID==""{computerError(w,422,"Cannot reach agent through its LAN addresses: "+last.Error());return}
 a.mu.Lock()
 for _,existing:=range a.cfg.Computers{
  if existing.URL==rec.URL||existing.Fingerprint==rec.Fingerprint {
   a.mu.Unlock();computerError(w,409,"This computer is already paired");return
  }
 }
 a.cfg.Computers=append(a.cfg.Computers,rec)
 if err:=a.save();err!=nil{
  a.cfg.Computers=a.cfg.Computers[:len(a.cfg.Computers)-1]
  a.mu.Unlock();computerError(w,500,"Cannot save computer connection");return
 }
 a.mu.Unlock()
 a.computerInviteMu.Lock()
 inv.State="connected";inv.ComputerID=rec.ID;inv.Message="Securely connected over "+rec.URL
 a.computerInviteMu.Unlock()
 success=true
 computerJSON(w,201,map[string]any{"connected":true,"computer_id":rec.ID,"address":rec.URL})
}
// The enrollment token is not itself the agent key. The agent generates and
// retains its own random key and private TLS certificate.
func computerEnrollmentHubURL(raw string)error{
 u,err:=url.Parse(raw)
 if err!=nil{return err}
 if u.Scheme!="https"||u.Hostname()==""||u.User!=nil||u.Path!=""||u.RawQuery!=""||u.Fragment!=""{
  return errors.New("Syncria enrollment requires an HTTPS origin with a trusted certificate")
 }
 return nil
}
func agentCertificatePin(raw []byte)string{
 sum:=sha256.Sum256(raw);return hex.EncodeToString(sum[:])
}
