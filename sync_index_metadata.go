package main

import "time"

// indexedPairItem records the exact metadata observed after a file is
// successfully reconciled. Never call it for an incomplete transfer.
func indexedPairItem(l,r bridgeEntry)Item {
 return Item{
  ID:l.ID,MD5:l.Fingerprint,LocalMD5:r.Fingerprint,Folder:l.Folder,
  IndexVersion:1,SourceSize:l.Size,DestinationSize:r.Size,
  SourceModified:l.Modified.UTC(),DestinationModified:r.Modified.UTC(),
 }
}
func indexedTimeEqual(a,b time.Time)bool{
 if a.IsZero()||b.IsZero(){return false}
 gap:=a.Sub(b)
 return gap<time.Second && gap > -time.Second
}

// During migration, old JSON checkpoints lack timestamps and use the
// previously persisted metadata fingerprints. No payload bytes are opened.
func metadataChanged(prior Item,l,r bridgeEntry)(bool,bool){
 if prior.IndexVersion>=1&&!prior.SourceModified.IsZero()&&!prior.DestinationModified.IsZero()&&
 !l.Modified.IsZero()&&!r.Modified.IsZero(){
  return prior.SourceSize!=l.Size||!indexedTimeEqual(prior.SourceModified,l.Modified),
   prior.DestinationSize!=r.Size||!indexedTimeEqual(prior.DestinationModified,r.Modified)
 }
 return prior.MD5!=l.Fingerprint||(prior.ID!=""&&l.ID!=""&&prior.ID!=l.ID),
  prior.LocalMD5!=r.Fingerprint
}
