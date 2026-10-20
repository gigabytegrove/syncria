package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/gigabytegrove/syncria/agentclient"
)

const version = "1.0.0-dev"

func run(args []string, input io.Reader, output io.Writer) error {
	if len(args) > 0 && args[0] == "grants" {
		return runGrants(args[1:], output)
	}
	if len(args) == 0 {
		return errors.New("usage: syncria-agent enroll|run|status [options]")
	}
	path, err := agentclient.DefaultFile()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	data := fs.String("data", path, "absolute identity file path")
	server := fs.String("server", "", "Syncria server URL (HTTPS)")
	once := fs.Bool("once", false, "send one heartbeat and exit")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	client := agentclient.Client{}
	switch args[0] {
	case "enroll":
		if *server == "" {
			return errors.New("enroll requires -server https://your-syncria-server")
		}
		if _, err := os.Lstat(*data); err == nil {
			return errors.New("agent already enrolled; saved credentials will not be overwritten")
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if _, err := agentclient.ValidServer(*server); err != nil {
			return err
		}
		fmt.Fprintln(output, "Enter the pairing token issued in Syncria, then press Enter:")
		token, err := bufio.NewReader(io.LimitReader(input, 256)).ReadString('\n')
		if err != nil && !(errors.Is(err, io.EOF) && token != "") {
			return err
		}
		token = strings.TrimSpace(token)
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		identity, err := client.Enroll(ctx, *server, token, runtime.GOOS, runtime.GOARCH, version)
		if err != nil {
			return err
		}
		if err := agentclient.Save(*data, identity); err != nil {
			return fmt.Errorf("server enrolled device but local credentials could not be saved: %w", err)
		}
		fmt.Fprintf(output, "Enrolled %s (%s). Device identity saved privately.\n", identity.Name, identity.ID)
		return nil
	case "status":
		identity, err := agentclient.Read(*data)
		if err != nil {
			return err
		}
		fmt.Fprintf(output, "Device: %s\nID: %s\nServer: %s\nEnrolled: %s\n", identity.Name, identity.ID, identity.Server, identity.EnrolledAt.Format(time.RFC3339))
		return nil
	case "run":
		identity, err := agentclient.Read(*data)
		if err != nil {
			return err
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		backoff := time.Second * 5
		for {
			heartbeatCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
			err := client.Heartbeat(heartbeatCtx, identity, runtime.GOOS, runtime.GOARCH, version)
			cancel()
			if err == nil {
				fmt.Fprintf(output, "%s: heartbeat accepted for %s\n", time.Now().Format(time.RFC3339), identity.ID)
				backoff = 5 * time.Second
			}
			if errors.Is(err, agentclient.ErrRevoked) {
				return fmt.Errorf("agent authentication rejected: %w", err)
			}
			if err != nil {
				log.Printf("heartbeat retry: %v", err)
			}
			if *once {
				return err
			}
			pause := time.Minute
			if err != nil {
				pause = backoff
				backoff *= 2
				if backoff > time.Minute {
					backoff = time.Minute
				}
			}
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(pause):
			}
		}
	default:
		return fmt.Errorf("unknown agent command %q; choose enroll, run, or status", args[0])
	}
}
func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
