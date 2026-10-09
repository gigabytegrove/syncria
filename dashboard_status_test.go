package main

import (
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
)

func TestDashboardStatusRejectsAnonymousAccess(t *testing.T) {
 a:=&App{}
 req:=httptest.NewRequest(http.MethodGet,"/api/dashboard/status",nil)
 w:=httptest.NewRecorder()
 a.dashboardStatus(w,req)
 if w.Code!=http.StatusUnauthorized {t.Fatalf("anonymous request returned %d",w.Code)}
}

func TestDashboardStatusIsGetOnly(t *testing.T) {
 a:=&App{}
 req:=httptest.NewRequest(http.MethodPost,"/api/dashboard/status",nil)
 w:=httptest.NewRecorder()
 a.dashboardStatus(w,req)
 if w.Code!=http.StatusMethodNotAllowed {t.Fatalf("POST returned %d",w.Code)}
}

func TestDashboardUIUpdatesElementsWithoutNavigation(t *testing.T) {
 for _,needed:=range []string{"/api/dashboard/status","data-job-enabled=","data-job-summary=","data-job-error=","data-share-mounted=","id=\"installedVersion\"","refreshDashboardStatus"} {
  if !strings.Contains(newUI,needed) {t.Errorf("missing live field: %s",needed)}
 }
 for _,forbidden:=range []string{"location.replace(","location.reload(","window.location.reload(","setInterval(checkRunningVersion"} {
  if strings.Contains(newUI,forbidden) {t.Errorf("automatic navigation detected: %s",forbidden)}
 }
}
