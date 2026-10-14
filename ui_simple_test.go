package main

import (
 "html/template"
 "os"
 "strings"
 "testing"
)

func TestSimpleUIRetainsCoreFeatures(t *testing.T){
 checks:=[]string{
  `id="syncWizard"`,`id="relationshipType"`,`id="sourceProvider"`,
  `id="nasSourceId"`,`id="nasTargets"`,`id="scheduleMode"`,
  `id="syncInterval"`,`id="nasDirection"`,`id="conflictPolicy"`,
  `name="cross_nas_propagation"`,`name="propagate_deletions" value="1" checked`,
  `name="recovery_enabled" value="1" checked`,
  `<option value="one_way">`,`<option value="bidirectional">`,
  `<option value="manual">`,`<option value="newer">`,`<option value="primary">`,
  `<option value="secondary">`,`<option value="keep_both">`,
  `/api/recovery/undo`,`/api/conflicts/resolve`,`/api/update/progress`,
  `/api/nas/action`,`/api/browse/dropbox`,
 }
 for _,part:=range checks{
  if !strings.Contains(newUI,part){t.Errorf("required feature/setting missing: %s",part)}
 }
}
func TestSimpleUILanguageAndGuidedSetup(t *testing.T){
 expected:=[]string{
  "New here? Start with these three steps.",
  "A NAS is a storage box",
  "Where are your files now?",
  "Where should copies go?",
  "What happens when I delete a file?",
  "Give me 10 minutes to undo an accidental deletion",
  "Ask me which copy to keep",
  "Save my sync (keep it off)",
  "My NAS Boxes",
  "My Syncs",
 }
 for _,phrase:=range expected{if !strings.Contains(newUI,phrase){t.Errorf("missing beginner text: %q",phrase)}}
 for _,phrase:=range []string{"Update staged. Syncria is restarting; refresh shortly.","<option value=\"overwrite_all\">"}{
  if strings.Contains(newUI,phrase){t.Errorf("obsolete or unapproved option returned: %q",phrase)}
 }
}
func TestSimpleUIAndThemeParse(t *testing.T){
 if _,err:=template.New("ui").Parse(newUI);err!=nil{t.Fatal(err)}
 if !strings.Contains(v051Theme,".quick-start-grid") || !strings.Contains(v051Theme,".glossary-layout") || !strings.Contains(v051Theme,"@media(max-width:800px)") {
  t.Fatal("guided setup and mobile help styles are missing")
 }
}
func TestFamilyDocumentationAvailable(t *testing.T){
 for _,file:=range []string{"README.md","docs/GETTING-STARTED.md","docs/RELEASE-v0.5.2.md","docs/TECHNICAL.md"}{
  b,err:=os.ReadFile(file);if err!=nil{t.Fatal(err)}
  if len(b)<300{t.Errorf("%s documentation appears incomplete",file)}
 }
}
