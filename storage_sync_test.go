package main

import (
    "context"
    "net/http"
    "net/http/httptest"
    "os"
    "path/filepath"
    "strings"
    "testing"
)

func TestStorageRelationshipCopiesAndTracksBothDirections(t *testing.T) {
    root:=t.TempDir()
    t.Setenv("SYNCRIA_STORAGE_ROOT",root)
    source,target:=filepath.Join(root,"source"),filepath.Join(root,"target")
    for _,p:=range []string{source,target}{if err:=os.Mkdir(p,0700);err!=nil{t.Fatal(err)}}
    if err:=os.WriteFile(filepath.Join(source,"a.txt"),[]byte("hello"),0600);err!=nil{t.Fatal(err)}
    j:=Job{ID:"test",Provider:"storage",RemoteID:source,Local:target,State:map[string]Item{}}
    a:=&App{}
    if err:=a.executeStorage(context.Background(),&j);err!=nil{t.Fatal(err)}
    data,err:=os.ReadFile(filepath.Join(target,"a.txt"))
    if err!=nil||string(data)!="hello"{t.Fatalf("source not copied: %q, %v",data,err)}
    if j.TotalUploads!=1||j.TotalUploadBytes!=5||j.TotalDownloads!=0{t.Fatalf("wrong counters: %+v",j)}
    if err:=os.WriteFile(filepath.Join(target,"a.txt"),[]byte("new value"),0600);err!=nil{t.Fatal(err)}
    if err:=a.executeStorage(context.Background(),&j);err!=nil{t.Fatal(err)}
    data,err=os.ReadFile(filepath.Join(source,"a.txt"))
    if err!=nil||string(data)!="new value"{t.Fatalf("reverse update failed: %q, %v",data,err)}
    if j.TotalDownloads!=1||j.TotalDownloadBytes!=9{t.Fatalf("reverse totals wrong: %+v",j)}
}

func TestStorageInitialConflictsAreNotOverwritten(t *testing.T){
    root:=t.TempDir()
    t.Setenv("SYNCRIA_STORAGE_ROOT",root)
    src,dst:=filepath.Join(root,"s"),filepath.Join(root,"d")
    for _,p:=range []string{src,dst}{if err:=os.Mkdir(p,0700);err!=nil{t.Fatal(err)}}
    _=os.WriteFile(filepath.Join(src,"x.txt"),[]byte("left"),0600)
    _=os.WriteFile(filepath.Join(dst,"x.txt"),[]byte("right"),0600)
    j:=Job{Provider:"storage",RemoteID:src,Local:dst,State:map[string]Item{}}
    err:=(&App{}).executeStorage(context.Background(),&j)
    if err==nil||!strings.Contains(err.Error(),"conflict"){t.Fatalf("expected safe conflict, got %v",err)}
    b,_:=os.ReadFile(filepath.Join(dst,"x.txt"))
    if string(b)!="right"{t.Fatal("destination unexpectedly overwritten")}
}

func TestStorageOverlappingPathsRejected(t *testing.T){
    root:=t.TempDir(); t.Setenv("SYNCRIA_STORAGE_ROOT",root)
    src:=filepath.Join(root,"source")
    if err:=os.MkdirAll(filepath.Join(src,"nested"),0700);err!=nil{t.Fatal(err)}
    if !pathsOverlap(src,filepath.Join(src,"nested")){t.Fatal("nested paths accepted")}
    if pathsOverlap(filepath.Join(root,"one"),filepath.Join(root,"two")){t.Fatal("distinct paths overlap")}
    j:=Job{Provider:"storage",RemoteID:src,Local:filepath.Join(src,"nested"),State:map[string]Item{}}
    if err:=(&App{}).executeStorage(context.Background(),&j);err==nil{t.Fatal("overlap accepted")}
}

func TestStorageDeletionRequiresReview(t *testing.T){
    root:=t.TempDir();t.Setenv("SYNCRIA_STORAGE_ROOT",root)
    src,dst:=filepath.Join(root,"s"),filepath.Join(root,"d")
    _=os.Mkdir(src,0700);_=os.Mkdir(dst,0700)
    _=os.WriteFile(filepath.Join(src,"a.txt"),[]byte("keep"),0600)
    j:=Job{Provider:"storage",RemoteID:src,Local:dst,State:map[string]Item{}}
    a:=&App{}
    if err:=a.executeStorage(context.Background(),&j);err!=nil{t.Fatal(err)}
    _=os.Remove(filepath.Join(src,"a.txt"))
    err:=a.executeStorage(context.Background(),&j)
    if err==nil||!strings.Contains(err.Error(),"Deletion"){t.Fatalf("expected deletion warning, got %v",err)}
    if _,err:=os.Stat(filepath.Join(dst,"a.txt"));err!=nil{t.Fatal("deleted opposite side")}
}

func TestMetricsPersistableTotals(t *testing.T){
    j:=Job{}
    recordTransfer(&j,true,100)
    recordTransfer(&j,false,200)
    j.SuccessfulRuns=2
    j.FailedRuns=1
    j.State=map[string]Item{"x.txt":{LocalMD5:"hash"},"folder":{Folder:true}}
    m:=summarizeMetrics([]Job{j})
    if m.Runs!=3||m.Failures!=1||m.UploadBytes!=100||m.DownloadBytes!=200||m.TrackedFiles!=1{t.Fatalf("bad stats: %+v",m)}
}

func TestWizardEndpointsRejectAnonymousRequests(t *testing.T){
    a:=&App{cfg:Config{Secret:b64rand()}}
    for _,handler:=range []http.HandlerFunc{a.verifyWizard,a.verifyNetworkStorage}{
        req:=httptest.NewRequest(http.MethodPost,"http://localhost/api/wizard/verify",strings.NewReader("kind=storage&path=/"))
        req.Header.Set("Content-Type","application/x-www-form-urlencoded")
        rec:=httptest.NewRecorder()
        handler(rec,req)
        if rec.Code!=http.StatusUnauthorized{t.Fatalf("expected 401, got %d",rec.Code)}
    }
}
