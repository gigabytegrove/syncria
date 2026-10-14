package main

import (
 "strings"
 "testing"
)
func TestSystemWideLiveUI(t *testing.T){
 required:=[]string{
  `id="liveSystemBar"`,
  `id="liveSystemDetail"`,
  `id="liveActivityFeed"`,
  `new EventSource('/api/live')`,
  `function applyLiveStatus(data)`,
  `beginLiveFallback()`,
 }
 for _,item:=range required{if !strings.Contains(newUI,item){t.Errorf("Missing live status feature %s",item)}}
 if !strings.Contains(v051Theme,".live-activity-item"){t.Fatal("Missing live activity styling")}
}
