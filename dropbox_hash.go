package main

import (
 "crypto/sha256"
 "encoding/hex"
 "hash"
)

// Dropbox's content hash is SHA256(SHA256(4MiB block 1) || ...).
// Only hashing state is kept in RAM, never a full file or local disk cache.
type dropboxHash struct {
 block hash.Hash
 outer hash.Hash
 count int
}
const dropboxBlockSize=4*1024*1024
func newDropboxHash()*dropboxHash{return &dropboxHash{block:sha256.New(),outer:sha256.New()}}
func (h *dropboxHash) Write(p []byte)(int,error){
 written:=len(p)
 for len(p)>0{
  remain:=dropboxBlockSize-h.count
  n:=len(p);if n>remain{n=remain}
  if _,err:=h.block.Write(p[:n]);err!=nil{return written-len(p),err}
  h.count+=n
  p=p[n:]
  if h.count==dropboxBlockSize{
   _,_=h.outer.Write(h.block.Sum(nil))
   h.block.Reset()
   h.count=0
  }
 }
 return written,nil
}
func (h *dropboxHash) HexSum()string{
 outer:=h.outer.(interface{Sum([]byte)[]byte}).Sum(nil)
 // If there is a partial final block, append its block digest to a
 // cloned outer hash without mutating the writer's state.
 if h.count==0{return hex.EncodeToString(outer)}
 clone:=sha256.New()
 // outer is SHA256 and cannot be cloned through hash.Hash; reconstruct
 // a block-wise hash from the serialized state instead of buffering data.
 state,err:=h.outer.(interface{MarshalBinary()([]byte,error)}).MarshalBinary()
 if err==nil{
  if err=clone.(interface{UnmarshalBinary([]byte)error}).UnmarshalBinary(state);err==nil{
   _,_=clone.Write(h.block.Sum(nil))
   return hex.EncodeToString(clone.Sum(nil))
  }
 }
 return ""
}
