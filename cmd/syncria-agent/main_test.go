package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gigabytegrove/syncria/agentclient"
)

func TestStatusAndInvalidCommands(t *testing.T) {
	var output bytes.Buffer
	if err := run([]string{}, strings.NewReader(""), &output); err == nil {
		t.Fatal("missing command accepted")
	}
	if err := run([]string{"bogus"}, strings.NewReader(""), &output); err == nil {
		t.Fatal("unknown command accepted")
	}
	path := filepath.Join(t.TempDir(), "identity.json")
	if err := agentclient.Save(path, agentclient.Credentials{ID: "device1", Name: "My Device", Server: "https://example.com", Secret: strings.Repeat("x", 43)}); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"status", "-data", path}, strings.NewReader(""), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "My Device") || strings.Contains(output.String(), strings.Repeat("x", 43)) {
		t.Fatal("status output incorrect or leaked credential")
	}
	if err := run([]string{"enroll", "-data", path, "-server", "https://example.com"}, strings.NewReader("token\n"), &output); err == nil {
		t.Fatal("reenrollment overwrote active identity")
	}
}
