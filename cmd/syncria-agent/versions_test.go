package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gigabytegrove/syncria/agentclient"
	"github.com/gigabytegrove/syncria/agentpolicy"
)

func TestVersionCommandsLocalEndToEnd(t *testing.T) {
	home := t.TempDir()
	identity := filepath.Join(home, "identity.json")
	token := "YXh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHg"
	if err := agentclient.Save(identity, agentclient.Credentials{Server: "https://syncria.example", ID: "device-1", Name: "Test Agent", Secret: token, EnrolledAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(home, "root")
	if err := os.MkdirAll(filepath.Join(root, "photos"), 0700); err != nil {
		t.Fatal(err)
	}
	live := filepath.Join(root, "photos", "notes.txt")
	if err := os.WriteFile(live, []byte("first version"), 0600); err != nil {
		t.Fatal(err)
	}
	grantsFile := filepath.Join(home, "grants.json")
	grants, err := agentpolicy.Open(grantsFile)
	if err != nil {
		t.Fatal(err)
	}
	g, err := grants.Add("Home", root, false, []string{"photos"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(home, "versions")
	common := []string{"-identity", identity, "-grants", grantsFile, "-archive", archive}
	var out bytes.Buffer
	args := append(append([]string{"save"}, common...), "-job", "aabbccddeeff00112233445566778899", "-grant", g.ID, "-relative", "photos/notes.txt")
	if err := runVersions(args, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Historical version saved") {
		t.Fatal("save did not report version")
	}
	out.Reset()
	args = append(append([]string{"list"}, common...), "-job", "aabbccddeeff00112233445566778899")
	if err := runVersions(args, &out); err != nil {
		t.Fatal(err)
	}
	fields := strings.Fields(out.String())
	if len(fields) < 3 || len(fields[0]) != 32 {
		t.Fatalf("bad history output: %q", out.String())
	}
	id := fields[0]
	if err := os.WriteFile(live, []byte("second version - preserve"), 0600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	args = append(append([]string{"restore"}, common...), "-version", id, "-grant", g.ID)
	if err := runVersions(args, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Existing files were not overwritten") {
		t.Fatalf("unsafe feedback: %q", out.String())
	}
	recovered := filepath.Join(root, "photos", "notes (restored-"+id[:12]+").txt")
	body, err := os.ReadFile(recovered)
	if err != nil || string(body) != "first version" {
		t.Fatalf("bad historical recovery: %v", err)
	}
	current, err := os.ReadFile(live)
	if err != nil || string(current) != "second version - preserve" {
		t.Fatal("current data was overwritten")
	}
	if err := runVersions(args, &out); !errors.Is(err, os.ErrExist) && !strings.Contains(err.Error(), "exists") {
		t.Fatalf("repeat restore replaced previous file: %v", err)
	}
}
func TestVersionCommandsRequireEnrolledIdentity(t *testing.T) {
	err := runVersions([]string{"list", "-identity", filepath.Join(t.TempDir(), "missing.json")}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "not enrolled") {
		t.Fatalf("unauthorized local invocation: %v", err)
	}
}
