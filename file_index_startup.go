package main

import (
 "errors"
 "fmt"
 "os"
 "path/filepath"
)

// initializeFileIndex is the exact startup migration path, kept testable
// without starting a server. Existing database contents are authoritative.
// If a previously indexed installation loses its DB, DO NOT rebaseline.
func (a *App) initializeFileIndex(dir string)error{
 indexPath:=filepath.Join(dir,"syncria-index.db")
 if a.cfg.FileIndexVersion>=1 {
  if _,err:=os.Stat(indexPath);err!=nil{
   return fmt.Errorf("Indexed state exists but database is missing or inaccessible; refusing to rebaseline: %w",err)
  }
 }
 // The original settings+pair checkpoint is backed up before conversion.
 // Create once; never overwrite the migration rollback copy.
 if a.cfg.FileIndexVersion==0 {
  if oldData,err:=os.ReadFile(a.file);err==nil{
   backup:=a.file+".pre-index-v0.8.0"
   if _,err:=os.Stat(backup);errors.Is(err,os.ErrNotExist){
    tmp:=backup+".tmp"
    if err=os.WriteFile(tmp,oldData,0600);err!=nil{return fmt.Errorf("Cannot create index migration backup: %w",err)}
    if err=os.Rename(tmp,backup);err!=nil{return fmt.Errorf("Cannot finalize index migration backup: %w",err)}
   }else if err!=nil{return fmt.Errorf("Cannot check previous backup: %w",err)}
  }else if !errors.Is(err,os.ErrNotExist){return fmt.Errorf("Cannot read old settings for backup: %w",err)}
 }
 idx,err:=loadSyncIndex(indexPath)
 if err!=nil{return err}
 if a.cfg.FileIndexVersion>=1{
  if err=idx.requireIndexedJobs(a.cfg.Jobs);err!=nil{idx.Close();return err}
 }
 if err=idx.hydratePairs(a.cfg.Jobs);err!=nil{idx.Close();return fmt.Errorf("Index import failed: %w",err)}
 a.index=idx
 if err=a.save();err!=nil{
  a.index=nil;idx.Close()
  return fmt.Errorf("Index migration state could not be committed: %w",err)
 }
 return nil
}
