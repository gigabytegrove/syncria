package main

import (
 "net/http"
 "strings"
)

func (a *App) conflictAction(w http.ResponseWriter,r *http.Request){
 if !localPost(r){http.Error(w,"Same-origin POST required",405);return}
 if !a.authenticated(r){http.Error(w,"Unauthorized",401);return}
 r.Body=http.MaxBytesReader(w,r.Body,8192)
 if err:=r.ParseForm();err!=nil{http.Error(w,"Invalid form",400);return}
 id:=r.FormValue("id");key:=r.FormValue("pair");rel:=r.FormValue("path")
 selection:=r.FormValue("choice")
 if selection!="primary"&&selection!="secondary"&&selection!="keep_both"{
  http.Error(w,"Choose the primary file, the secondary file, or keep both",400);return
 }
 if id==""||key==""||rel==""||strings.ContainsRune(rel,0){http.Error(w,"Invalid conflict",400);return}
 if _,running:=a.runs.Load(id);running{redirect(w,r,"Wait for the current synchronization to finish");return}
 a.mu.Lock()
 job:=a.job(id)
 if job==nil{a.mu.Unlock();http.Error(w,"Relationship not found",404);return}
 exists:=false
 for _,conflict:=range job.Conflicts{if conflict.Pair==key&&conflict.Path==rel{exists=true;break}}
 if !exists{a.mu.Unlock();http.Error(w,"No unresolved conflict matches the requested file",404);return}
 if job.ConflictOverrides==nil{job.ConflictOverrides=map[string]string{}}
 job.ConflictOverrides[key+"|"+rel]=selection
 err:=a.save()
 a.mu.Unlock()
 if err!=nil{http.Error(w,"Could not save your conflict choice",500);return}
 redirect(w,r,"Conflict choice saved. Run this relationship to apply the selected resolution.")
}
