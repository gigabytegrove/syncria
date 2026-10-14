package main

import (
 "encoding/base64"
 "strings"
 "regexp"
)

const portableDrivePrefix = "~syncria~"
const duplicateDriveMarker = "~syncria-id~"
var validDriveID = regexp.MustCompile(`^[A-Za-z0-9_-]{5,}package main

import (
 "encoding/base64"
 "strings"
 "regexp"
)

)

// Google Drive permits literal slash and backslash characters in filenames.
// SMB/NAS filesystems don't. Encode the entire component (not a path) when
// required, and also encode names already starting with our reserved marker
// so every Drive name maps to a unique, reversible NAS component.
func portableDriveName(name string) string {
 if strings.ContainsAny(name,"/\\") || strings.HasPrefix(strings.ToLower(name),portableDrivePrefix) || strings.Contains(strings.ToLower(name),duplicateDriveMarker) {
  return portableDrivePrefix+base64.RawURLEncoding.EncodeToString([]byte(name))
 }
 return name
}
func portableDriveObjectName(name,id string,duplicate bool) string {
 base:=portableDriveName(name)
 if duplicate && validDriveID.MatchString(id){return base+duplicateDriveMarker+id}
 return base
}
func originalDriveName(component string) string {
 if i:=strings.LastIndex(component,duplicateDriveMarker);i>=0 && validDriveID.MatchString(component[i+len(duplicateDriveMarker):]) {component=component[:i]}

 if !strings.HasPrefix(strings.ToLower(component),portableDrivePrefix){return component}
 payload:=component[len(portableDrivePrefix):]
 raw,err:=base64.RawURLEncoding.DecodeString(payload)
 if err!=nil{return component}
 name:=string(raw)
 if portableDriveName(name)!=component{return component}
 return name
}
