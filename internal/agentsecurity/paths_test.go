package agentsecurity

import (
 "os"
 "path/filepath"
 "runtime"
 "testing"
)

func TestAuthorizedPath(t *testing.T){
 root:=t.TempDir()
 if err:=os.Mkdir(filepath.Join(root,"allowed"),0700);err!=nil{t.Fatal(err)}
 out,err:=AuthorizedPath(root,filepath.Join("allowed","new.txt"))
 if err!=nil{t.Fatal(err)}
 if out!=filepath.Join(root,"allowed","new.txt"){t.Fatalf("unexpected result %q",out)}
 for _,bad:=range []string{"","..",filepath.Join("..","escape"),filepath.Join(root,"absolute")}{
  if _,err:=AuthorizedPath(root,bad);err==nil{t.Errorf("accepted unsafe path %q",bad)}
 }
}

func TestRejectSymlinkTraversal(t *testing.T){
 root:=t.TempDir();elsewhere:=t.TempDir()
 if err:=os.Symlink(elsewhere,filepath.Join(root,"outside"));err!=nil{
  if runtime.GOOS=="windows"{t.Skipf("symlink privilege unavailable: %v",err)}
  t.Fatal(err)
 }
 if _,err:=AuthorizedPath(root,filepath.Join("outside","file"));err==nil{t.Fatal("symlink escape accepted")}
}
