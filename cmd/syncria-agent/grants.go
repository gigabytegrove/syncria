package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/gigabytegrove/syncria/agentclient"
	"github.com/gigabytegrove/syncria/agentpolicy"
)

type folderSelectors []string

func (s *folderSelectors) String() string { return strings.Join(*s, ",") }
func (s *folderSelectors) Set(v string) error {
	if v == "" {
		return errors.New("empty selector")
	}
	*s = append(*s, v)
	return nil
}

// grants are device-owner-only local operations; there is intentionally no
// controller endpoint that can create or broaden a local folder permission.
func runGrants(args []string, output io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: syncria-agent grants add|list|remove [flags]")
	}
	identityPath, err := agentclient.DefaultFile()
	if err != nil {
		return err
	}
	sub := args[0]
	fs := flag.NewFlagSet("grants "+sub, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	state := fs.String("state", filepath.Join(filepath.Dir(identityPath), "grants.json"), "private local grant database")
	root := fs.String("root", "", "absolute existing directory (add)")
	name := fs.String("name", "", "human-readable folder name (add)")
	readonly := fs.Bool("read-only", true, "deny remote write operations by policy")
	id := fs.String("id", "", "grant ID (remove)")
	var only, exclude folderSelectors
	fs.Var(&only, "select", "relative included folder, repeat as needed (add)")
	fs.Var(&exclude, "exclude", "relative excluded folder, repeat as needed (add)")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return errors.New("unexpected arguments")
	}
	if !filepath.IsAbs(*state) {
		return errors.New("grant database path must be absolute")
	}
	switch sub {
	case "add", "list", "remove":
	default:
		return errors.New("unknown grants command: choose add, list or remove")
	}
	store, err := agentpolicy.Open(*state)
	if err != nil {
		return err
	}
	switch sub {
	case "add":
		if *root == "" || *name == "" {
			return errors.New("grant add requires -root and -name")
		}
		g, err := store.Add(*name, *root, *readonly, only, exclude)
		if err != nil {
			return err
		}
		fmt.Fprintf(output, "Approved folder %s (%s)\nPath: %s\nRead-only: %t\nSelections: %d\nExclusions: %d\n", g.Label, g.ID, g.Root, g.ReadOnly, len(g.Selections), len(g.Exclusions))
	case "list":
		grants := store.List()
		if len(grants) == 0 {
			fmt.Fprintln(output, "No local folders approved.")
			return nil
		}
		for _, g := range grants {
			fmt.Fprintf(output, "%s (%s)\n  Directory: %s\n  Read-only: %t\n", g.Label, g.ID, g.Root, g.ReadOnly)
			for _, p := range g.Selections {
				fmt.Fprintf(output, "  Included: %s\n", p)
			}
			for _, p := range g.Exclusions {
				fmt.Fprintf(output, "  Excluded: %s\n", p)
			}
		}
	case "remove":
		if *id == "" {
			return errors.New("grant remove requires -id")
		}
		if err := store.Remove(*id); err != nil {
			return err
		}
		fmt.Fprintf(output, "Removed local folder permission %s.\n", *id)
	}
	return nil
}

// The approved roots themselves are never transmitted in the agent heartbeat.
// Future file adapters must authorize each root ID and verify operation-time
// filesystem handles before any scan/read/write is permitted.
