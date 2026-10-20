package agentsecurity

import (
 "errors"
 "io/fs"
 "os"
 "path/filepath"
 "strings"
)

// AuthorizedPath confines a requested relative path to an approved root.
// Symlink traversal is rejected, including the final component. Callers
// must additionally use handle-relative filesystem APIs to defeat TOCTOU
// races before accepting untrusted concurrent filesystem modifications.
func AuthorizedPath(root,relative string)(string,error){
 if root==""||!filepath.IsAbs(root){return "",errors.New("approved root must be absolute")}
 if relative==""||filepath.IsAbs(relative){return "",errors.New("path must be relative and nonempty")}
 if strings.ContainsRune(relative,0){return "",errors.New("invalid path")}
 clean:=filepath.Clean(relative)
 if clean=="."||clean==".."||strings.HasPrefix(clean,".."+string(filepath.Separator)){return "",errors.New("path escapes approved root")}
 resolvedRoot,err:=filepath.EvalSymlinks(root);if err!=nil{return "",err}
 stat,err:=os.Stat(resolvedRoot);if err!=nil{return "",err}
 if !stat.IsDir(){return "",errors.New("approved root is not a directory")}
 candidate:=filepath.Join(resolvedRoot,clean)
 rel,err:=filepath.Rel(resolvedRoot,candidate)
 if err!=nil||rel==".."||strings.HasPrefix(rel,".."+string(filepath.Separator)){return "",errors.New("path escapes approved root")}
 current:=resolvedRoot
 for _,part:=range strings.Split(rel,string(filepath.Separator)){
  current=filepath.Join(current,part)
  info,err:=os.Lstat(current)
  if errors.Is(err,fs.ErrNotExist){continue}
  if err!=nil{return "",err}
  if info.Mode()&os.ModeSymlink!=0{return "",errors.New("symlink traversal is forbidden")}
 }
 return candidate,nil
}
