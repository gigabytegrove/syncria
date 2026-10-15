package main

import (
 "html/template"
 "strings"
 "testing"
)

// The screen redesign must not silently remove essential synchronization
// controls, file transparency or recovery. Keep these assertions focused on
// real workflows, not decorative CSS implementation details.
func TestCohesiveSyncWorkspacePreservesFullWorkflow(t *testing.T){
 if _,err:=template.New("syncria-workspace").Parse(newUI);err!=nil{t.Fatalf("HTML template invalid: %v",err)}
 for _,required:=range []string{
  `id="syncLibrarySearch"`,`id="syncLibraryFilter"`,`data-sync-card="{{.ID}}"`,
  `onclick="openSyncDetails('{{.ID}}')"`,
  `id="syncDetailOverview"`,`id="syncDetailFiles"`,`id="syncDetailHistory"`,`id="syncDetailSettings"`,
  `role="tablist"`,`aria-selected="true"`,
  `id="syncDetailPreviewHost"`,`id="syncRunEvents"`,`id="syncFileSearch"`,`id="syncFileActionFilter"`,
  `name="action" value="run_job"`,`name="action" value="toggle_job"`,`name="action" value="delete_job"`,
  `requestStopSync(`,`editSync(`,`previewSync(`,`data-preview-result="`,`/api/conflicts/resolve`,`/api/recovery/undo`,
  `id="syncDeleteWarning"`,`id="syncDetailStatus"`,
 } {
  if !strings.Contains(newUI,required){t.Errorf("Missing essential flow %s",required)}
 }
}

func TestCohesiveUIStylesAreStandaloneAndResponsive(t *testing.T){
 for _,required:=range []string{
  ".sync-library-header",".sync-toolbar-search",".sync-card-more-menu",
  ".sync-details-tabs",".sync-details-tab.is-active",".sync-detail-pairs",
  ".sync-run-history-layout",".sync-history-file-reason",
  "@media(max-width:640px)","prefers-reduced-motion",
 } {
  if !strings.Contains(syncriaDesignCSS,required){t.Errorf("Missing UX styling: %s",required)}
 }
 if !strings.Contains(newUI,"v051Theme + syncriaDesignCSS"){t.Fatal("Workspace stylesheet is not embedded after the existing design system")}
}

func TestRunHistoryFiltersAreServerSideAndEveryFileRemainsSearchable(t *testing.T){
 for _,needed:=range []string{
  `oninput="debounceRunFileSearch()"`,`onchange="reloadRunFileFilters()"`,
  `const data=await fetchSyncHistory({id,run:runID,offset:eventPageOffset,q:query,action})`,
  `Show`,`runDetailPreview()`,
 } {
  if !strings.Contains(newUI,needed){t.Errorf("File history is not fully interactive: %s",needed)}
 }
 if !strings.Contains(newUI,"unmountSyncPreview()"){t.Fatal("Preview reuse needs DOM-safe unmount behavior")}
}
