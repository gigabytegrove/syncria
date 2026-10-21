package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/gigabytegrove/syncria/agentclient"
	"github.com/gigabytegrove/syncria/agentpolicy"
	"github.com/gigabytegrove/syncria/agentrestore"
	"github.com/gigabytegrove/syncria/agentversion"
)

// Version operations are local-only and require an enrolled device identity.
// The controller never receives file bytes, and recovery refuses overwrite.
func runVersions(args []string, output io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: syncria-agent versions save|list|restore [options]")
	}
	defaultIdentity, err := agentclient.DefaultFile()
	if err != nil {
		return err
	}
	configDir := filepath.Dir(defaultIdentity)
	fs := flag.NewFlagSet("versions "+args[0], flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	identity := fs.String("identity", defaultIdentity, "private agent identity")
	grantsPath := fs.String("grants", filepath.Join(configDir, "grants.json"), "approved local folder policy")
	archive := fs.String("archive", filepath.Join(configDir, "versions"), "private destination-owned archive")
	quota := fs.Int64("quota-bytes", 64<<30, "maximum historical version bytes")
	job := fs.String("job", "", "existing sync job ID")
	grant := fs.String("grant", "", "locally approved folder ID")
	relative := fs.String("relative", "", "path within approved folder")
	id := fs.String("version", "", "saved version identifier")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return errors.New("unexpected version arguments")
	}
	if _, err := agentclient.Read(*identity); err != nil {
		return fmt.Errorf("local agent not enrolled: %w", err)
	}
	versions, err := agentversion.Open(*archive, *quota)
	if err != nil {
		return err
	}
	switch args[0] {
	case "list":
		for _, v := range versions.List(*job, *relative) {
			fmt.Fprintf(output, "%s  %s  %s  %d bytes  %s\n", v.ID, v.JobID, v.Relative, v.Bytes, v.Created.Format("2006-01-02T15:04:05Z07:00"))
		}
		return nil
	case "save", "restore":
		grants, err := agentpolicy.Open(*grantsPath)
		if err != nil {
			return err
		}
		manager := agentrestore.Manager{Grants: grants, Versions: versions}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		if args[0] == "save" {
			if *job == "" || *grant == "" || *relative == "" {
				return errors.New("versions save requires -job -grant -relative")
			}
			v, err := manager.Save(ctx, *job, *grant, *relative)
			if err != nil {
				return err
			}
			fmt.Fprintf(output, "Historical version saved on this device: %s (%d bytes, SHA-256 %s).\n", v.ID, v.Bytes, v.SHA256)
			return nil
		}
		if *id == "" || *grant == "" {
			return errors.New("versions restore requires -version and -grant")
		}
		result, err := manager.Restore(ctx, *id, *grant)
		if err != nil {
			return err
		}
		fmt.Fprintf(output, "Recovered %s (%d bytes) into approved folder. Existing files were not overwritten.\n", result.Path, result.Bytes)
		return nil
	default:
		return fmt.Errorf("unknown versions command %q", args[0])
	}
}
