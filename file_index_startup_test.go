package main

import (
 "encoding/json"
 "os"
 "path/filepath"
 "testing"
)

func TestIndexStartupMigratesThenRestartsWithoutLosingFiles(t *testing.T){
 dir:=t.TempDir()
 file:=filepath.Join(dir,"state.json")
 initial:=App{file:file,cfg:Config{Jobs:[]Job{{
  ID:"job-existing",Name:"Drive to NAS",Provider:"bridge",
  PairStates:map[string]map[string]Item{"account|nas|folder":{
   "Kaleb/video.mp4":{ID:"cloud-id",MD5:"cloud-metadata",LocalMD5:"nas-metadata"},
  }},
 }}}}
 if err:=initial.save();err!=nil{t.Fatal(err)}
 original,err:=os.ReadFile(file);if err!=nil{t.Fatal(err)}
 if err=initial.initializeFileIndex(dir);err!=nil{t.Fatal(err)}
 if err=initial.index.Close();err!=nil{t.Fatal(err)}
 backup,err:=os.ReadFile(file+".pre-index-v0.8.0")
 if err!=nil{t.Fatalf("No rollback backup: %v",err)}
 if string(backup)!=string(original){t.Fatal("Original checkpoint was not safely backed up")}
 migrated,err:=os.ReadFile(file);if err!=nil{t.Fatal(err)}
 var cfg Config
 if err=json.Unmarshal(migrated,&cfg);err!=nil{t.Fatal(err)}
 if cfg.FileIndexVersion!=1{t.Fatal("Indexed configuration missing version marker")}
 if len(cfg.Jobs[0].PairStates)!=0{t.Fatal("Large file records were not migrated out of state.json")}
 a:=App{file:file,cfg:cfg}
 if err=a.initializeFileIndex(dir);err!=nil{t.Fatalf("Restart failed: %v",err)}
 if a.cfg.Jobs[0].PairStates["account|nas|folder"]["Kaleb/video.mp4"].ID!="cloud-id"{t.Fatal("Restart did not restore indexed file")}
 if err=a.index.Close();err!=nil{t.Fatal(err)}
}

func TestIndexStartupRefusesMissingDatabaseAfterMigration(t *testing.T){
 dir:=t.TempDir()
 file:=filepath.Join(dir,"state.json")
 a:=App{file:file,cfg:Config{Jobs:[]Job{{ID:"existing",Name:"Important sync"}}}}
 if err:=a.save();err!=nil{t.Fatal(err)}
 if err:=a.initializeFileIndex(dir);err!=nil{t.Fatal(err)}
 if err:=a.index.Close();err!=nil{t.Fatal(err)}
 persisted,err:=os.ReadFile(file);if err!=nil{t.Fatal(err)}
 if err=os.Remove(filepath.Join(dir,"syncria-index.db"));err!=nil{t.Fatal(err)}
 var cfg Config
 if err=json.Unmarshal(persisted,&cfg);err!=nil{t.Fatal(err)}
 restart:=App{file:file,cfg:cfg}
 if err=restart.initializeFileIndex(dir);err==nil{t.Fatal("Missing DB after migration must not trigger a new baseline")}
}
