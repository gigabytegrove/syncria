package main

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestGrantCommandsDurableAndSelective(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "pictures")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	state := filepath.Join(base, "data", "grants.json")
	var output bytes.Buffer
	args := []string{"grants", "add", "-state", state, "-root", root, "-name", "My photos", "-select", "family", "-exclude", "family/private"}
	if err := run(args, strings.NewReader(""), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Selections: 1") || !strings.Contains(output.String(), "Read-only: true") {
		t.Fatal(output.String())
	}
	id := regexp.MustCompile(`\(([a-f0-9]{32})\)`).FindStringSubmatch(output.String())
	if len(id) != 2 {
		t.Fatal(output.String())
	}
	output.Reset()
	if err := run([]string{"grants", "list", "-state", state}, nil, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "family/private") || !strings.Contains(output.String(), id[1]) {
		t.Fatal(output.String())
	}
	output.Reset()
	if err := run([]string{"grants", "remove", "-state", state, "-id", id[1]}, nil, &output); err != nil {
		t.Fatal(err)
	}
	output.Reset()
	if err := run([]string{"grants", "list", "-state", state}, nil, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "No local folders approved") {
		t.Fatal(output.String())
	}
}
func TestGrantCommandsRejectInvalidAndUnsafeInputs(t *testing.T) {
	for _, args := range [][]string{
		{"grants"}, {"grants", "unknown"}, {"grants", "add"}, {"grants", "add", "-name", "n", "-root", "/not-existing-anywhere"},
		{"grants", "remove"}, {"grants", "add", "-state", "relative", "-root", "/tmp", "-name", "n"},
	} {
		if err := run(args, nil, &bytes.Buffer{}); err == nil {
			t.Errorf("accepted invalid arguments: %q", args)
		}
	}
}
