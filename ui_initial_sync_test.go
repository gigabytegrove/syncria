package main

import (
 "strings"
 "testing"
)

func TestInitialCloudSyncPreferenceIsVisibleAndExplained(t *testing.T){
 for _,want:=range []string{
  `id="cloudFirstRunFields"`,
  `name="initial_conflict_policy"`,
  `value="newer">Use the file with the newer modification time`,
  `value="manual">Ask me which copy to keep`,
  `row.primary_modified`,
  `row.secondary_modified`,
  `First run: Syncria proposes a newer copy`,
 }{
  if !strings.Contains(newUI,want){t.Errorf("Missing first-run control or date evidence: %s",want)}
 }
}
