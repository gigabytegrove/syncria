package main

import (
 "context"
 "crypto/md5"
 "encoding/hex"
 "errors"
 "fmt"
 "io"
 "net"
 "os"
 "path"
 "strconv"
 "strings"
 "time"

 smb2 "github.com/cloudsoda/go-smb2"
)

// directSMB is a remote SMB2/3 connection, NOT a kernel mount. File contents
// remain on the NAS. Only bounded io.Copy buffers exist within Syncria.
type directSMB struct {
 socket net.Conn
 session *smb2.Session
 share *smb2.Share
}

func safeRemotePath(folder string) (string,error) {
 folder=strings.ReplaceAll(folder,"\\","/")
 if folder==""||folder=="."||folder=="/"{return ".",nil}
 if strings.HasPrefix(folder,"/") || strings.ContainsRune(folder,0){return "",errors.New("Remote paths must be relative to the selected share")}
 p:=path.Clean(folder)
 if p==".."||strings.HasPrefix(p,"../"){return "",errors.New("Path escapes the NAS share")}
 for _,part:=range strings.Split(p,"/"){
  if !safeName(part)||strings.ContainsAny(part,";\n\r"){return "",fmt.Errorf("Unsafe NAS folder name %q",part)}
 }
 return p,nil
}
func joinRemotePath(root,rel string)(string,error){
 root,err:=safeRemotePath(root);if err!=nil{return "",err}
 if rel=="" {return root,nil}
 rel,err=safeRemotePath(rel);if err!=nil{return "",err}
 joined:=path.Join(root,rel)
 if root!="."&&joined!=root&&!strings.HasPrefix(joined,root+"/"){return "",errors.New("Remote path escaped selected folder")}
 return joined,nil
}

// These are existing Syncria credentials stored in a private 0600 file in
// /data/connections, never in a URL or process argument.
func (a *App) smbCredentials(id string)(string,string,error){
 b,err:=os.ReadFile(credsFile(a,id));if err!=nil{return "","",fmt.Errorf("NAS credentials not available: %w",err)}
 var username,password string
 for _,line:=range strings.Split(string(b),"\n"){
  if strings.HasPrefix(line,"username="){username=strings.TrimPrefix(line,"username=")}
  if strings.HasPrefix(line,"password="){password=strings.TrimPrefix(line,"password=")}
 }
 if username==""||password==""{return "","",errors.New("NAS credentials are incomplete")}
 return username,password,nil
}

func (a *App) dialSMB(ctx context.Context,s StorageShare)(*directSMB,error){
 if s.Protocol!="smb"{return nil,errors.New("Only direct SMB2/3 connections are implemented; NFS requires a native protocol connector")}
 if err:=validateShare(s.Protocol,s.Server,s.Export,s.Label);err!=nil{return nil,err}
 // Only local/private IPs: do not allow the control UI to connect to public
 // services using the storage connector.
 host,err:=validateDiscoveryIP(s.Server);if err!=nil{return nil,err}
 username,password,err:=a.smbCredentials(s.ID);if err!=nil{return nil,err}
 target:=net.JoinHostPort(host,strconv.Itoa(445))
 d:=&smb2.Dialer{Initiator:&smb2.NTLMInitiator{User:username,Password:password}}
 session,err:=d.Dial(ctx,target)
 if err!=nil{return nil,fmt.Errorf("SMB authentication failed or port 445 unavailable: %w",err)}
 fs,err:=session.Mount(s.Export)
 if err!=nil{session.Logoff();return nil,fmt.Errorf("SMB share unavailable: %w",err)}
 return &directSMB{session:session,share:fs.WithContext(ctx)},nil
}
func (s *directSMB) Close(){
 if s.share!=nil{_ =s.share.Umount()}
 if s.session!=nil{_ =s.session.Logoff()}
 if s.socket!=nil{_ =s.socket.Close()}
}
type smbObject struct {
 File bool
 Size int64
 Fingerprint string
 Modified time.Time
}
func (s *directSMB) scan(ctx context.Context,root string)(map[string]smbObject,error){
 root,err:=safeRemotePath(root);if err!=nil{return nil,err}
 found:=map[string]smbObject{}
 var walk func(dir,rel string,depth int)error
 walk=func(dir,rel string,depth int)error{
  if depth>64{return errors.New("NAS directory depth exceeded")}
  if err:=ctx.Err();err!=nil{return err}
  infos,err:=s.share.ReadDir(dir)
  if err!=nil{return fmt.Errorf("List NAS folder %q: %w",dir,err)}
  seen:=map[string]bool{}
  for _,info:=range infos {
   name:=info.Name()
   // Internal recovery files remain on the NAS for Undo and are never
   // treated as ordinary files or a fatal scan error.
   if name==".syncria-recovery" {continue}
   if !safeName(name){return fmt.Errorf("Unsafe NAS item name %q",name)}
   folded:=strings.ToLower(name);if seen[folded]{return fmt.Errorf("Case-colliding NAS filenames under %q",dir)}
   seen[folded]=true
   if info.Mode()&os.ModeSymlink!=0{return fmt.Errorf("NAS symbolic link %q not supported",name)}
   childRel:=path.Join(rel,name)
   child,err:=joinRemotePath(root,childRel);if err!=nil{return err}
   if len(found)>100000{return errors.New("NAS listing exceeds 100,000 objects")}
   found[childRel]=smbObject{File:!info.IsDir(),Size:info.Size(),Fingerprint:fmt.Sprintf("%d:%d",info.Size(),info.ModTime().UnixNano()),Modified:info.ModTime()}
   if info.IsDir(){if err:=walk(child,childRel,depth+1);err!=nil{return err}}
  }
  return nil
 }
 return found,walk(root,"",0)
}
func (s *directSMB) mkdir(root,rel string)error{
 target,err:=joinRemotePath(root,rel);if err!=nil{return err}
 if target=="."{return nil}
 parts:=strings.Split(target,"/")
 for i:=range parts{
  p:=strings.Join(parts[:i+1],"/")
  err=s.share.Mkdir(p,0750)
  if err!=nil && !os.IsExist(err) {
   stat,statErr:=s.share.Stat(p)
   if statErr!=nil||!stat.IsDir(){return fmt.Errorf("Create NAS folder %q: %w",p,err)}
  }
 }
 return nil
}
func (s *directSMB) open(root,rel string)(io.ReadCloser,int64,error){
 file,err:=joinRemotePath(root,rel);if err!=nil{return nil,0,err}
 f,err:=s.share.Open(file);if err!=nil{return nil,0,err}
 stat,err:=f.Stat()
 if err!=nil{f.Close();return nil,0,err}
 if !stat.Mode().IsRegular(){f.Close();return nil,0,errors.New("NAS item is not a regular file")}
 return f,stat.Size(),nil
}

// put streams source -> NAS temporary file. The partial file is only on
// the NAS, never the machine or container hosting Syncria.
func (s *directSMB) put(ctx context.Context,root,rel string,source io.Reader,expectedSize int64,expectedMD5 string,report func(int64))error {
 dest,err:=joinRemotePath(root,rel);if err!=nil{return err}
 if err=s.mkdir(root,path.Dir(rel));err!=nil{return err}
 temp:=dest+".syncria-pending-"+newID()
 f,err:=s.share.Create(temp);if err!=nil{return err}
 complete:=false
 defer func(){if !complete{_ =s.share.Remove(temp)}}()
 const chunk=128*1024
 buf:=make([]byte,chunk)
 var n int64
 checksum:=md5.New()
 dropboxChecksum:=newDropboxHash()
 for {
  if err:=ctx.Err();err!=nil{f.Close();return err}
  count,readErr:=source.Read(buf)
  if count>0{
   written,writeErr:=f.Write(buf[:count]);n+=int64(written)
   if writeErr!=nil{f.Close();return writeErr}
   if written!=count{f.Close();return io.ErrShortWrite}
   if _,e:=checksum.Write(buf[:written]);e!=nil{f.Close();return e}
   if _,e:=dropboxChecksum.Write(buf[:written]);e!=nil{f.Close();return e}
   if report!=nil{report(n)}
  }
  if readErr==io.EOF{break}
  if readErr!=nil{f.Close();return readErr}
 }
 if err=f.Close();err!=nil{return err}
 if expectedSize>=0&&expectedSize!=n{return fmt.Errorf("Transfer length mismatch: expected %d, wrote %d",expectedSize,n)}
 if expectedMD5!=""{
  switch len(expectedMD5){
  case 32:
   if !strings.EqualFold(hex.EncodeToString(checksum.Sum(nil)),expectedMD5){return errors.New("Source MD5 mismatch; original NAS file preserved")}
  case 64:
   if !strings.EqualFold(dropboxChecksum.HexSum(),expectedMD5){return errors.New("Dropbox source content hash mismatch; original NAS file preserved")}
  default:return errors.New("Unrecognized source checksum; NAS destination unchanged")
  }
 }
 // Renaming a completed staging file is performed on the NAS. Existing
 // destination files are never removed first; if the NAS disallows
 // replacement the operation fails safely rather than deleting user data.
 previous:=""
 if _,e:=s.share.Stat(dest);e==nil {
  previous=dest+".syncria-old-"+newID()
  if err=s.share.Rename(dest,previous);err!=nil{return fmt.Errorf("Cannot safely preserve existing NAS file: %w",err)}
 }
 if err=s.share.Rename(temp,dest);err!=nil{
  if previous!="" {
   if restoreErr:=s.share.Rename(previous,dest);restoreErr!=nil{return fmt.Errorf("NAS commit failed and previous file could not be restored at %s: %v / %v",dest,err,restoreErr)}
  }
  return fmt.Errorf("NAS commit failed; original preserved: %w",err)
 }
 if previous!="" { _=s.share.Remove(previous) }
 complete=true
 return nil
}
