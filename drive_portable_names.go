package main

import (
 "encoding/base64"
 "strings"
)

const portableDrivePrefix = "~syncria~"

// Google Drive permits literal slash and backslash characters in filenames.
// SMB/NAS filesystems don't. Encode the entire component (not a path) when
// required, and also encode names already starting with our reserved marker
// so every Drive name maps to a unique, reversible NAS component.
func portableDriveName(name string) string {
 if strings.ContainsAny(name,"/\\") || strings.HasPrefix(strings.ToLower(name),portableDrivePrefix) {
  return portableDrivePrefix+base64.RawURLEncoding.EncodeToString([]byte(name))
 }
 return name
}
func originalDriveName(component string) string {
 if !strings.HasPrefix(strings.ToLower(component),portableDrivePrefix){return component}
 payload:=component[len(portableDrivePrefix):]
 raw,err:=base64.RawURLEncoding.DecodeString(payload)
 if err!=nil{return component}
 name:=string(raw)
 if portableDriveName(name)!=component{return component}
 return name
}
