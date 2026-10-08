package main

import (
 "embed"
 "net/http"
)

//go:embed assets/syncria-logo.svg
var brandAssets embed.FS

func serveBrandLogo(w http.ResponseWriter,r *http.Request){
 if r.URL.Path!="/assets/syncria-logo.svg"{http.NotFound(w,r);return}
 data,err:=brandAssets.ReadFile("assets/syncria-logo.svg")
 if err!=nil {http.Error(w,"Brand logo unavailable",http.StatusInternalServerError);return}
 w.Header().Set("Content-Type","image/svg+xml")
 w.Header().Set("Cache-Control","public, max-age=86400")
 w.Header().Set("X-Content-Type-Options","nosniff")
 _,_=w.Write(data)
}
