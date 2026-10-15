package main

import (
 "strings"
 "testing"
)

func TestIndexedDatabaseBlocksBlindRollback(t *testing.T){
 a:=&App{cfg:Config{FileIndexVersion:1}}
 err:=a.beginUpdate("rollback")
 if err==nil{t.Fatal("Old executable rollback should not be allowed after index migration")}
 if !strings.Contains(err.Error(),"state.json.pre-index-v0.8.0"){t.Fatalf("Rollback blocker lacks recovery guidance: %v",err)}
 if a.updateBusy{t.Fatal("Refused rollback must not mark updater busy")}
}
