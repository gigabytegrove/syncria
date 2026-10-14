package main

import (
 "context"
 "io"
 "strings"
 "testing"
 "time"
)

type firstRunMemoryEndpoint struct{
 bytes string
 reads int
}
func (e *firstRunMemoryEndpoint) Scan(context.Context)(map[string]bridgeEntry,error){return nil,nil}
func (e *firstRunMemoryEndpoint) Mkdir(context.Context,string)(bridgeEntry,error){return bridgeEntry{},nil}
func (e *firstRunMemoryEndpoint) Open(_ context.Context,_ string)(io.ReadCloser,int64,error){
 e.reads++
 return io.NopCloser(strings.NewReader(e.bytes)),int64(len(e.bytes)),nil
}
func (e *firstRunMemoryEndpoint) Put(context.Context,string,io.Reader,int64,string,func(int64))(bridgeEntry,error){return bridgeEntry{},nil}
func (e *firstRunMemoryEndpoint) Close(){}

func TestFirstRunTimestampAvoidsAllFileReads(t *testing.T){
 ctx:=context.Background()
 stamp:=time.Date(2026,10,14,12,14,0,0,time.UTC)
 huge:=int64(30)*1024*1024*1024
 cloud:=&firstRunMemoryEndpoint{bytes:"source"}
 nas:=&firstRunMemoryEndpoint{bytes:"destination"}

 matched,err:=bridgeSameFirstRun(ctx,cloud,nas,"movie.mov",
  bridgeEntry{Size:huge,Modified:stamp},
  bridgeEntry{Size:huge,Modified:stamp})
 if err!=nil||!matched{t.Fatalf("Matching timestamps and sizes must skip file bytes: equal=%v err=%v",matched,err)}
 if cloud.reads!=0||nas.reads!=0{t.Fatal("Metadata check opened file contents")}

 matched,err=bridgeSameFirstRun(ctx,cloud,nas,"movie.mov",
  bridgeEntry{Size:huge,Modified:stamp.Add(time.Second)},
  bridgeEntry{Size:huge,Modified:stamp})
 if err!=nil||matched{t.Fatalf("Newer file must require transfer decision: equal=%v err=%v",matched,err)}
 if cloud.reads!=0||nas.reads!=0{t.Fatal("Timestamp comparison opened file contents")}

 matched,err=bridgeSameFirstRun(ctx,cloud,nas,"movie.mov",
  bridgeEntry{Size:huge,Modified:stamp},
  bridgeEntry{Size:huge-1,Modified:stamp})
 if err==nil||matched{t.Fatal("Equal timestamps and different sizes must require review")}
 if cloud.reads!=0||nas.reads!=0{t.Fatal("Size mismatch opened file contents")}
}
