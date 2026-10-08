package main

import (
 "context"
 "encoding/json"
 "errors"\n "io"
 "net/http"
 "net/url"
 "os"
 "path/filepath"
 "runtime"
 "sort"
 "strings"
 "time"
)

type browseEntry struct {Name string `json:"name"`; ID string `json:"id"`; Kind string `json:"kind"`}
type browseResponse struct {Path string `json:"path"`; Parent string `json:"parent,omitempty"`; Entries []browseEntry `json:"entries"`; Message string `json:"message,omitempty"`}

func browseError(w http.ResponseWriter,code int,msg string){http.Error(w,msg,code)}
func sendBrowse(w http.ResponseWriter,v browseResponse){
 w.Header().Set("Content-Type","application/json; charset=utf-8")
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("X-Content-Type-Options","nosniff")
 _=json.NewEncoder(w).Encode(v)
}
func browseStorageRoot() string {
 root:=os.Getenv("SYNCRIA_STORAGE_ROOT")
 if root=="" {if runtime.GOOS=="windows" {root=filepath.VolumeName(os.Getenv("SystemDrive"))+"\\"}else{root="/storage";if _,err:=os.Stat(root);err!=nil{root="/"}}}
 abs,err:=filepath.Abs(root);if err!=nil{return root};return filepath.Clean(abs)
}
func contained(root,target string) bool {
 r,err:=filepath.Rel(root,target)
 return err==nil && r!=".." && !strings.HasPrefix(r,".."+string(os.PathSeparator))
}
func resolvedPath(path string)(string,error){
 p,err:=filepath.EvalSymlinks(path);if err!=nil{return "",err}
 return filepath.Abs(p)
}
func (a *App) browseLocal(w http.ResponseWriter,r *http.Request){
 if r.Method!="GET"{browseError(w,405,"GET required");return}
 if !a.authenticated(r){browseError(w,401,"Unauthorized");return}
 root:=browseStorageRoot();physicalRoot,err:=resolvedPath(root);if err!=nil{browseError(w,503,"Storage root unavailable: "+err.Error());return}
 chosen:=r.URL.Query().Get("path");if chosen==""{chosen=root}
 if !filepath.IsAbs(chosen){browseError(w,400,"Absolute path required");return}
 target,err:=resolvedPath(filepath.Clean(chosen))
 if err!=nil{browseError(w,404,"Folder unavailable: "+err.Error());return}
 if !contained(physicalRoot,target){browseError(w,403,"Outside configured storage root");return}
 f,err:=os.Open(target);if err!=nil{browseError(w,403,"Cannot open folder: "+err.Error());return};defer f.Close()
 stat,err:=f.Stat();if err!=nil||!stat.IsDir(){browseError(w,400,"Not a directory");return}
 entries,err:=f.ReadDir(1001);if err!=nil && !errors.Is(err,io.EOF) {
  if len(entries)==0 {browseError(w,403,"Cannot read folder: "+err.Error());return}
 }
 res:=browseResponse{Path:target,Entries:[]browseEntry{}}
 if target!=physicalRoot{res.Parent=filepath.Dir(target)}
 for _,entry:=range entries{
  name:=entry.Name()
  if name=="."||name==".."{continue}
  full:=filepath.Join(target,name)
  st,e:=os.Stat(full);if e!=nil||!st.IsDir(){continue}
  real,e:=resolvedPath(full);if e!=nil||!contained(physicalRoot,real){continue}
  res.Entries=append(res.Entries,browseEntry{Name:name,ID:real,Kind:"folder"})
 }
 sort.Slice(res.Entries,func(i,j int)bool{return strings.ToLower(res.Entries[i].Name)<strings.ToLower(res.Entries[j].Name)})
 if len(entries)>1000{res.Message="Showing first 1000 entries; narrow the folder"}
 sendBrowse(w,res)
}
func (a *App) browseDrive(w http.ResponseWriter,r *http.Request){
 if r.Method!="GET"{browseError(w,405,"GET required");return}
 if !a.authenticated(r){browseError(w,401,"Unauthorized");return}
 acc:=r.URL.Query().Get("account")
 parent:=r.URL.Query().Get("parent");if parent==""{parent="root"}
 if parent!="root"&&!validDriveID(parent){browseError(w,400,"Invalid Drive folder ID");return}
 a.mu.Lock();exists:=a.account(acc)!=nil;a.mu.Unlock()
 if !exists{browseError(w,400,"Select a connected Google account");return}
 ctx,cancel:=context.WithTimeout(r.Context(),45*time.Second);defer cancel()
 children,err:=a.folderChildren(ctx,acc,parent);if err!=nil{browseError(w,502,"Google Drive: "+err.Error());return}
 result:=browseResponse{Path:parent,Entries:[]browseEntry{}}
 for _,file:=range children {
  if file.MimeType==folderMime && safeName(file.Name) {
   result.Entries=append(result.Entries,browseEntry{Name:file.Name,ID:file.ID,Kind:"folder"})
  }
 }
 sort.Slice(result.Entries,func(i,j int)bool{return strings.ToLower(result.Entries[i].Name)<strings.ToLower(result.Entries[j].Name)})
 if parent!="root"{
  q:=url.Values{};q.Set("fields","id,parents")
  var info struct{Parents []string `json:"parents"`}
  if err:=a.apiJSON(ctx,acc,"GET","https://www.googleapis.com/drive/v3/files/"+url.PathEscape(parent)+"?"+q.Encode(),nil,&info);err==nil && len(info.Parents)>0{result.Parent=info.Parents[0]}else{result.Parent="root"}
 }
 sendBrowse(w,result)
}
