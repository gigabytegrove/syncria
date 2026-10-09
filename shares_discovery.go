package main

import (
 "context"
 "encoding/json"
 "errors"
 "fmt"
 "net"
 "net/http"
 "os"
 "os/exec"
 "sort"
 "strings"
 "time"
)

func validateDiscoveryIP(server string) (string,error) {
 ip:=net.ParseIP(strings.TrimSpace(server))
 if ip==nil||!(ip.IsPrivate()||ip.IsLoopback()||ip.IsLinkLocalUnicast()) {return "",errors.New("Enter a private/local IP address")}
 // Link-local scopes need an explicit interface and are not supported here.
 return ip.String(),nil
}
func discoverSMBShares(ctx context.Context,server string)([]string,error){
 return discoverSMBSharesAuthenticated(ctx,server,"","")
}
func discoverSMBSharesAuthenticated(ctx context.Context,server,username,password string)([]string,error){
 host,err:=validateDiscoveryIP(server);if err!=nil{return nil,err}
 if _,err=exec.LookPath("smbclient");err!=nil{return nil,errors.New("SMB discovery tool missing from this runtime image; rebuild the image with samba-client")}
 ctx,cancel:=context.WithTimeout(ctx,10*time.Second);defer cancel()
 args:=[]string{"-L","//"+host,"-g"}
 if username=="" {args=append(args,"-N")}else{
  // Credential files prevent exposing passwords in process arguments.
  file,err:=os.CreateTemp("","syncria-smb-discover-*");if err!=nil{return nil,err}
  name:=file.Name();defer os.Remove(name)
  if err=file.Chmod(0600);err!=nil{file.Close();return nil,err}
  if strings.ContainsAny(username,"\n\r")||strings.ContainsAny(password,"\n\r"){file.Close();return nil,errors.New("Invalid SMB credentials")}
  _,err=fmt.Fprintf(file,"username = %s\npassword = %s\n",username,password)
  closeErr:=file.Close();if err!=nil{return nil,err};if closeErr!=nil{return nil,closeErr}
  args=append(args,"-A",name)
 }
 output,err:=exec.CommandContext(ctx,"smbclient",args...).CombinedOutput()
 if ctx.Err()!=nil{return nil,errors.New("SMB discovery timed out")}
 if err!=nil{return nil,errors.New("SMB share listing unavailable; check credentials or enter a known share")}
 found:=map[string]bool{}
 for _,line:=range strings.Split(string(output),"\n"){
  parts:=strings.Split(strings.TrimSpace(line),"|")
  if len(parts)<2||!strings.EqualFold(strings.TrimSpace(parts[0]),"Disk"){continue}
  name:=strings.TrimSpace(parts[1])
  if shareName.MatchString(name)&&!strings.HasSuffix(name,"$"){found[name]=true}
 }
 shares:=make([]string,0,len(found));for name:=range found{shares=append(shares,name)};sort.Strings(shares)
 return shares,nil
}
func discoverNFSExports(ctx context.Context,server string)([]string,error){
 host,err:=validateDiscoveryIP(server);if err!=nil{return nil,err}
 if _,err=exec.LookPath("showmount");err!=nil{return nil,errors.New("NFS discovery tool missing from runtime image; install nfs-utils")}
 c,cancel:=context.WithTimeout(ctx,10*time.Second);defer cancel()
 output,err:=exec.CommandContext(c,"showmount","-e",host).CombinedOutput()
 if c.Err()!=nil{return nil,errors.New("NFS export discovery timed out")}
 if err!=nil{return nil,errors.New("NFS exports not advertised over mountd; NFSv4-only servers may require a manual export path")}
 found:=map[string]bool{}
 for _,line:=range strings.Split(string(output),"\n"){
  fields:=strings.Fields(strings.TrimSpace(line))
  if len(fields)>0&&strings.HasPrefix(fields[0],"/")&&!strings.ContainsAny(fields[0],"\x00\r\n")&&!strings.Contains(fields[0],".."){found[fields[0]]=true}
 }
 shares:=make([]string,0,len(found));for name:=range found{shares=append(shares,name)};sort.Strings(shares)
 return shares,nil
}
func (a *App) shareDiscover(w http.ResponseWriter,r *http.Request){
 w.Header().Set("Content-Type","application/json; charset=utf-8")
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("X-Content-Type-Options","nosniff")
 if r.Method!=http.MethodPost{http.Error(w,"POST required",405);return}
 if !localPost(r){http.Error(w,"Same-origin request required",403);return}
 if !a.authenticated(r){http.Error(w,"Unauthorized",401);return}
 r.Body=http.MaxBytesReader(w,r.Body,8192)
 if err:=r.ParseForm();err!=nil{http.Error(w,"Invalid request",400);return}
 protocol:=strings.ToLower(strings.TrimSpace(r.FormValue("protocol")))
 server:=r.FormValue("server")
 var shares []string;var err error
 switch protocol{
 case "smb":shares,err=discoverSMBSharesAuthenticated(r.Context(),server,r.FormValue("username"),r.FormValue("password"))
 case "nfs":shares,err=discoverNFSExports(r.Context(),server)
 default:err=errors.New("Select SMB or NFS")
 }
 if err!=nil{w.WriteHeader(400);_=json.NewEncoder(w).Encode(map[string]string{"error":err.Error()});return}
 _=json.NewEncoder(w).Encode(struct{Shares []string `json:"shares"`}{shares})
}
