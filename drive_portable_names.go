package main

import (
 "encoding/base64"
 "strings"
)

const portableDrivePrefix = "~syncria~"
const duplicateDriveMarker = "~syncria-id~"

// Distinct Drive objects may share the same display name in one folder.
// Only those colliding siblings get stable IDs in their NAS filenames.
func portableDriveName(name string) string {
 if strings.ContainsAny(name,"/\\") || strings.HasPrefix(strings.ToLower(name),portableDrivePrefix) || strings.Contains(strings.ToLower(name),duplicateDriveMarker) {
  return portableDrivePrefix+base64.RawURLEncoding.EncodeToString([]byte(name))
 }
 return name
}
func portableDriveObjectName(name,id string,duplicate bool) string {
 base:=portableDriveName(name)
 if duplicate && id!="" {return base+duplicateDriveMarker+base64.RawURLEncoding.EncodeToString([]byte(id))}
 return base
}
func originalDriveName(component string) string {
 if i:=strings.LastIndex(component,duplicateDriveMarker);i>=0 {
  encodedID:=component[i+len(duplicateDriveMarker):]
  if id,err:=base64.RawURLEncoding.DecodeString(encodedID);err==nil&&len(id)>0&&strings.IndexByte(string(id),0)<0 {component=component[:i]}
 }
 if !strings.HasPrefix(strings.ToLower(component),portableDrivePrefix){return component}
 payload:=component[len(portableDrivePrefix):]
 raw,err:=base64.RawURLEncoding.DecodeString(payload)
 if err!=nil{return component}
 name:=string(raw)
 if portableDriveName(name)!=component{return component}
 return name
}
