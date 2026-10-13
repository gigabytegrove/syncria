package main

import "testing"

func TestExplicitCloudNASPrimaryMayAlsoServeAsNASSyncSource(t *testing.T){
 first:=NASDestination{ShareID:"nas1",Folder:"Families/Drive"}
 cloud:=Job{Provider:"bridge",NASDestinations:[]NASDestination{first}}
 downstream:=Job{Provider:"nas_bridge",NASSourceID:"nas1",NASSourceFolder:"Families/Drive",NASDestinations:[]NASDestination{{ShareID:"nas2",Folder:"FamilyMirror"}}}
 if !allowedDesignatedNASOverlap(cloud,downstream,first){
  t.Fatal("Cloud destination and NAS-to-NAS primary should be shareable")
 }
 if !allowedDesignatedNASOverlap(downstream,cloud,first){
  t.Fatal("Order of relationship creation must not disable designated source sharing")
 }
}
func TestOtherOverlappingNASFoldersStayProtected(t *testing.T){
 designated:=NASDestination{ShareID:"nas1",Folder:"Family"}
 additional:=NASDestination{ShareID:"nas2",Folder:"Mirror"}
 cloud:=Job{Provider:"bridge",NASDestinations:[]NASDestination{designated,additional}}
 downstream:=Job{Provider:"nas_bridge",NASSourceID:"nas1",NASSourceFolder:"Family"}
 for _,folder:=range []NASDestination{
  {ShareID:"nas1",Folder:"Family/Child"},
  {ShareID:"nas2",Folder:"Mirror"},
  {ShareID:"nas1",Folder:""},
 }{
  if allowedDesignatedNASOverlap(cloud,downstream,folder){
   t.Fatalf("Unexpected overlap accepted: %+v",folder)
  }
 }
}
func TestPerNASPairStateUsesConnectionAndFolder(t *testing.T){
 d1:=NASDestination{ShareID:"nas1",Folder:"A"}
 d2:=NASDestination{ShareID:"nas2",Folder:"A"}
 d3:=NASDestination{ShareID:"nas1",Folder:"B"}
 if relationPairKey("cloud-account",d1)==relationPairKey("cloud-account",d2)||
    relationPairKey("cloud-account",d1)==relationPairKey("cloud-account",d3){
  t.Fatal("Different NAS shares/folders cannot share transfer state")
 }
}
func TestCrossNASCloudOriginSuppressionRespectsDefault(t *testing.T){
 primary:=NASDestination{ShareID:"nas1",Folder:"."}
 secondary:=NASDestination{ShareID:"nas2",Folder:"."}
 j:=Job{Provider:"bridge",CloudProvider:"google_drive",AccountID:"account",RemoteID:"root",NASDestinations:[]NASDestination{primary,secondary}}
 firstKey:=relationPairKey("google_drive:account:root",primary)
 secondKey:=relationPairKey("google_drive:account:root",secondary)
 recordCloudOrigin(&j,firstKey,"changed.txt","f1")
 if !suppressCrossNASTransfer(&j,secondKey,"changed.txt",bridgeEntry{Fingerprint:"f1"},true,true){
  t.Fatal("Default disabled cross-NAS propagation was ignored")
 }
 if suppressCrossNASTransfer(&j,secondKey,"cloud-edited.txt",bridgeEntry{Fingerprint:"different"},true,true){
  t.Fatal("Cloud changes should still reach selected NAS destinations")
 }
 j.CrossNASPropagation=true
 if suppressCrossNASTransfer(&j,secondKey,"changed.txt",bridgeEntry{Fingerprint:"f1"},true,true){
  t.Fatal("Opt-in cross-NAS propagation did not override default suppression")
 }
}
