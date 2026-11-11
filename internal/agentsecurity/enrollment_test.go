package agentsecurity

import (
 "errors"
 "sync"
 "testing"
 "time"
)

func TestEnrollmentLifecycle(t *testing.T){
 r:=NewRegistry()
 e,code,err:=r.Request("Desktop",time.Minute);if err!=nil{t.Fatal(err)}
 if _,_,err=r.Redeem(code);!errors.Is(err,ErrNotApproved){t.Fatalf("expected approval barrier: %v",err)}
 if err=r.Approve(code);err!=nil{t.Fatal(err)}
 id,key,err:=r.Redeem(code);if err!=nil{t.Fatal(err)}
 if id!=e.ID||len(key)!=64{t.Fatal("unexpected identity")}
 if err=r.Authenticate(id,key);err!=nil{t.Fatal(err)}
 if err=r.Authenticate(id,"wrong");err==nil{t.Fatal("wrong secret accepted")}
 if _,_,err=r.Redeem(code);!errors.Is(err,ErrInvalidCode){t.Fatal("code replay accepted")}
 if err=r.Revoke(id);err!=nil{t.Fatal(err)}
 if err=r.Authenticate(id,key);!errors.Is(err,ErrRevoked){t.Fatalf("revoke failed: %v",err)}
}

func TestExpiryAndInvalidInput(t *testing.T){
 r:=NewRegistry()
 if _,_,err:=r.Request("",time.Minute);err==nil{t.Fatal("empty name accepted")}
 if _,_,err:=r.Request("ok",0);err==nil{t.Fatal("zero ttl accepted")}
 _,code,err:=r.Request("ok",time.Second);if err!=nil{t.Fatal(err)}
 r.now=func()time.Time{return time.Now().Add(time.Hour)}
 if err=r.Approve(code);!errors.Is(err,ErrInvalidCode){t.Fatalf("expired code approved: %v",err)}
}

func TestConcurrentRedeemOnlyOneSucceeds(t *testing.T){
 r:=NewRegistry();_,code,err:=r.Request("ok",time.Minute);if err!=nil{t.Fatal(err)}
 if err=r.Approve(code);err!=nil{t.Fatal(err)}
 var wg sync.WaitGroup
 results:=make(chan bool,16)
 for i:=0;i<16;i++{wg.Add(1);go func(){defer wg.Done();_,_,e:=r.Redeem(code);results<-e==nil}()}
 wg.Wait();close(results);successful:=0
 for ok:=range results{if ok{successful++}}
 if successful!=1{t.Fatalf("expected one successful redemption, got %d",successful)}
}
