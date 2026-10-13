package main

import (
 "bytes"
 "crypto/sha256"
 "encoding/hex"
 "testing"
)

func TestDropboxHashShortPayload(t *testing.T){
 content:=[]byte("hello world")
 block:=sha256.Sum256(content)
 expected:=sha256.Sum256(block[:])
 digest:=newDropboxHash()
 for _,chunk:=range [][]byte{content[:2],content[2:8],content[8:]}{
  if _,err:=digest.Write(chunk);err!=nil{t.Fatal(err)}
 }
 if digest.HexSum()!=hex.EncodeToString(expected[:]){t.Fatalf("Dropbox hash incorrect")}
}
func TestDropboxHashMultipleBlocksWithoutDataCache(t *testing.T){
 data:=bytes.Repeat([]byte("x"),dropboxBlockSize+100)
 first:=sha256.Sum256(data[:dropboxBlockSize])
 second:=sha256.Sum256(data[dropboxBlockSize:])
 outer:=sha256.New();outer.Write(first[:]);outer.Write(second[:])
 expected:=hex.EncodeToString(outer.Sum(nil))
 stream:=newDropboxHash()
 for offset:=0;offset<len(data);offset+=4097{
  end:=offset+4097;if end>len(data){end=len(data)}
  if _,err:=stream.Write(data[offset:end]);err!=nil{t.Fatal(err)}
 }
 if stream.HexSum()!=expected{t.Fatal("Dropbox 4MiB block boundary hash mismatch")}
}
func TestDropboxHashEmptyFile(t *testing.T){
 digest:=newDropboxHash()
 empty:=sha256.Sum256(nil)
 if digest.HexSum()!=hex.EncodeToString(empty[:]){t.Fatal("Empty content hash incorrect")}
}
