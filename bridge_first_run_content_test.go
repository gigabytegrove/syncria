package main

import (
 "context"
 "crypto/md5"
 "encoding/hex"
 "io"
 "strings"
 "testing"
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

func TestFirstRunChecksumAvoidsSecondCloudRead(t *testing.T){
 body:="the-same-content"
 digest:=md5.Sum([]byte(body))
 known:=hex.EncodeToString(digest[:])
 // A nil-content Drive endpoint intentionally has no working Open method:
 // the MD5 shortcut must only read the NAS side.
 cloud:=&bridgeDriveEndpoint{}
 nas:=&firstRunMemoryEndpoint{bytes:body}
 equal,err:=bridgeSameFirstRun(context.Background(),cloud,nas,"file.txt",bridgeEntry{
  Size:int64(len(body)),Fingerprint:known,
 },bridgeEntry{Size:int64(len(body)),Fingerprint:"not-a-hash"})
 if err!=nil||!equal||nas.reads!=1{t.Fatalf("Expected one NAS read, no Drive read: equal=%v err=%v NAS=%d",equal,err,nas.reads)}
 nas.reads=0
 equal,err=bridgeSameFirstRun(context.Background(),cloud,nas,"file.txt",bridgeEntry{
  Size:1000000,Fingerprint:known,
 },bridgeEntry{Size:int64(len(body))})
 if err!=nil||equal||nas.reads!=0{t.Fatal("Mismatched sizes should skip byte reads")}
}
