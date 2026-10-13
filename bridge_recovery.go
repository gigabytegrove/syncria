package main

import (
 "context"
 "errors"
 "fmt"
 "net/http"
 "net/url"
 "path"
 "strings"
)

// Remote deletion safeguards operate exclusively on NAS or cloud storage.
// Syncria persists only the token and expiration in relationship metadata.
type recoverableBridgeEndpoint interface {
 bridgeEndpoint
 StageDelete(context.Context,string,string)(string,error)
 Restore(context.Context,string,string) error
 Purge(context.Context,string) error
 Delete(context.Context,string) error
}

func (b *bridgeSMBEndpoint) StageDelete(ctx context.Context, rel, jobID string)(string,error){
 if err:=ctx.Err();err!=nil{return "",err}
 original,err:=joinRemotePath(b.root,rel);if err!=nil{return "",err}
 if strings.HasPrefix(original,".syncria-recovery/"){return "",errors.New("Recovery internals cannot be synchronized")}
 directory:=path.Join(".syncria-recovery",jobID)
 // Internal recovery names are never accepted as user-selected paths.
 // Create the reserved directories directly, outside the public browser.
 if err=b.client.share.Mkdir(".syncria-recovery",0700);err!=nil{
  if info,e:=b.client.share.Stat(".syncria-recovery");e!=nil||!info.IsDir(){return "",err}
 }
 if err=b.client.share.Mkdir(directory,0700);err!=nil{
  if info,e:=b.client.share.Stat(directory);e!=nil||!info.IsDir(){return "",err}
 }
 token:=path.Join(directory,newID())
 if err=b.client.share.Rename(original,token);err!=nil{return "",fmt.Errorf("Moving deleted NAS item to recovery: %w",err)}
 return token,nil
}
func (b *bridgeSMBEndpoint) Restore(ctx context.Context,rel,token string)error {
 if err:=ctx.Err();err!=nil{return err}
 original,err:=joinRemotePath(b.root,rel);if err!=nil{return err}
 if !strings.HasPrefix(token,".syncria-recovery/"){return errors.New("Invalid recovery token")}
 if _,err=b.client.share.Stat(original);err==nil{return fmt.Errorf("Cannot restore %s: destination already exists",rel)}
 if err=b.client.mkdir(b.root,path.Dir(rel));err!=nil{return err}
 return b.client.share.Rename(token,original)
}
func (b *bridgeSMBEndpoint) Purge(ctx context.Context,token string)error {
 if err:=ctx.Err();err!=nil{return err}
 if !strings.HasPrefix(token,".syncria-recovery/"){return errors.New("Invalid recovery token")}
 return b.client.removeRemoteRecursive(ctx,token)
}
func (b *bridgeSMBEndpoint) Delete(ctx context.Context,rel string)error {
 if err:=ctx.Err();err!=nil{return err}
 name,err:=joinRemotePath(b.root,rel);if err!=nil{return err}
 return b.client.removeRemoteRecursive(ctx,name)
}
func (c *directSMB) removeRemoteRecursive(ctx context.Context,target string)error{
 if err:=ctx.Err();err!=nil{return err}
 info,err:=c.share.Stat(target);if err!=nil{return err}
 if info.IsDir(){
  children,err:=c.share.ReadDir(target);if err!=nil{return err}
  for _,child:=range children {
   if !safeName(child.Name()){return errors.New("Unsafe NAS filename in deletion target")}
   if err=c.removeRemoteRecursive(ctx,path.Join(target,child.Name()));err!=nil{return err}
  }
 }
 return c.share.Remove(target)
}

func (b *bridgeDriveEndpoint) StageDelete(ctx context.Context, rel, jobID string)(string,error){
 file,ok:=b.files[rel];if !ok {return "",errors.New("Drive item no longer exists")}
 if err:=b.app.apiJSON(ctx,b.account,http.MethodPatch,
  "https://www.googleapis.com/drive/v3/files/"+url.PathEscape(file.ID),
  map[string]any{"trashed":true},nil);err!=nil{return "",err}
 return file.ID,nil
}
func (b *bridgeDriveEndpoint) Restore(ctx context.Context,rel,token string)error{
 if token==""||!validDriveID(token){return errors.New("Invalid Drive recovery token")}
 return b.app.apiJSON(ctx,b.account,http.MethodPatch,
  "https://www.googleapis.com/drive/v3/files/"+url.PathEscape(token),
  map[string]any{"trashed":false},nil)
}
func (b *bridgeDriveEndpoint) Purge(ctx context.Context,token string)error{
 if token==""||!validDriveID(token){return errors.New("Invalid Drive recovery token")}
 uri:="https://www.googleapis.com/drive/v3/files/"+url.PathEscape(token)
 var metadata struct{ID string `json:"id"`;Trashed bool `json:"trashed"`}
 if err:=b.app.apiJSON(ctx,b.account,http.MethodGet,uri+"?fields=id,trashed",nil,&metadata);err!=nil{return err}
 // A file restored by someone in Google Drive must NEVER be deleted
 // automatically after Syncria's 10-minute undo window expires.
 if !metadata.Trashed{return nil}
 resp,err:=b.app.drive(ctx,b.account,http.MethodDelete,uri,nil,"")
 if err!=nil{return err}
 defer resp.Body.Close()
 return nil
}
func (b *bridgeDriveEndpoint) Delete(ctx context.Context,rel string)error{
 f,ok:=b.files[rel];if !ok{return errors.New("Google Drive file not found")}
 return b.Purge(ctx,f.ID)
}
