package main

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "net"
    "net/http"
    "net/url"
    "strings"
    "time"
)

func wizardJSON(w http.ResponseWriter, code int, ok bool, message string) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.Header().Set("Cache-Control", "no-store")
    w.Header().Set("X-Content-Type-Options", "nosniff")
    w.WriteHeader(code)
    _ = json.NewEncoder(w).Encode(struct{
        OK bool `json:"ok"`
        Message string `json:"message"`
    }{ok,message})
}

func (a *App) verifyWizard(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost || !localPost(r) {wizardJSON(w,405,false,"Same-origin POST required");return}
    if !a.authenticated(r) {wizardJSON(w,401,false,"Sign in to verify folders");return}
    if err := r.ParseForm();err != nil {wizardJSON(w,400,false,"Invalid form");return}
    kind := r.FormValue("kind")
    switch kind {
    case "storage":
        path,err := a.checkedStoragePath(strings.TrimSpace(r.FormValue("path")))
        if err != nil {wizardJSON(w,400,false,err.Error());return}
        wizardJSON(w,200,true,"Storage folder accessible: "+path)
    case "drive":
        acc,folder := r.FormValue("account"),strings.TrimSpace(r.FormValue("folder"))
        if folder == "" {folder="root"}
        if folder!="root" && !validDriveID(folder) {wizardJSON(w,400,false,"Invalid Drive folder ID");return}
        a.mu.Lock();exists:=a.account(acc)!=nil;a.mu.Unlock()
        if !exists {wizardJSON(w,400,false,"Connect a Google account first");return}
        ctx,cancel:=context.WithTimeout(r.Context(),20*time.Second);defer cancel()
        // Verify the OAuth token, Drive permission and selected folder.
        if folder=="root" {
            if _,err:=a.folderChildren(ctx,acc,"root");err!=nil{wizardJSON(w,502,false,"Google Drive could not be verified: "+err.Error());return}
        } else {
            var info struct {ID string `json:"id"`;MimeType string `json:"mimeType"`;Trashed bool `json:"trashed"`}
            endpoint:="https://www.googleapis.com/drive/v3/files/"+url.PathEscape(folder)+"?fields=id,mimeType,trashed"
            if err:=a.apiJSON(ctx,acc,"GET",endpoint,nil,&info);err!=nil{wizardJSON(w,502,false,"Drive folder could not be verified: "+err.Error());return}
            if info.ID!=folder || info.MimeType!=folderMime || info.Trashed {wizardJSON(w,400,false,"Selected Drive folder is not available");return}
        }
        wizardJSON(w,200,true,"Google Drive access verified")
    default:
        wizardJSON(w,400,false,"Unknown verification type")
    }
}

// This checks protocol, input and server reachability. Mounting/authentication
// is separately verified by the existing connect operation.
func (a *App) verifyNetworkStorage(w http.ResponseWriter,r *http.Request){
    if r.Method!=http.MethodPost || !localPost(r){wizardJSON(w,405,false,"Same-origin POST required");return}
    if !a.authenticated(r){wizardJSON(w,401,false,"Sign in to verify connections");return}
    if err:=r.ParseForm();err!=nil{wizardJSON(w,400,false,"Invalid form");return}
    protocol,server,export,label:=strings.ToLower(strings.TrimSpace(r.FormValue("protocol"))),strings.TrimSpace(r.FormValue("server")),strings.TrimSpace(r.FormValue("export")),strings.TrimSpace(r.FormValue("label"))
    if err:=validateShare(protocol,server,export,label);err!=nil{wizardJSON(w,400,false,err.Error());return}
    port:="445"
    if protocol=="nfs"{port="2049"}
    ctx,cancel:=context.WithTimeout(r.Context(),6*time.Second);defer cancel()
    var dialer net.Dialer
    conn,err:=dialer.DialContext(ctx,"tcp",net.JoinHostPort(server,port))
    if err!=nil{
        msg:=fmt.Sprintf("%s server not reachable on TCP %s: %s",strings.ToUpper(protocol),port,err)
        wizardJSON(w,503,false,msg);return
    }
    if err:=conn.Close();err!=nil && !errors.Is(err,net.ErrClosed){ /* connection reached */ }
    wizardJSON(w,200,true,strings.ToUpper(protocol)+" server reachable. Final connection verifies mount and credentials.")
}
