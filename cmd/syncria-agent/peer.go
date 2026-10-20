package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gigabytegrove/syncria/agentclient"
	"github.com/gigabytegrove/syncria/agentpeer"
	"github.com/gigabytegrove/syncria/agentpolicy"
	"github.com/gigabytegrove/syncria/agentpublish"
	"github.com/gigabytegrove/syncria/agenttransfer"
)

// Local device-owner operations never pass tokens via command-line arguments.
// A peer's token is intentionally separate from its controller identity.
func runPeer(args []string, input io.Reader, output io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: syncria-agent peer init|serve|inventory|stage|publish|receive")
	}
	identityFile, err := agentclient.DefaultFile()
	if err != nil {
		return err
	}
	dir := filepath.Dir(identityFile)
	fs := flag.NewFlagSet("peer "+args[0], flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	identity := fs.String("identity", identityFile, "private device identity path")
	grantsPath := fs.String("grants", filepath.Join(dir, "grants.json"), "approved local folders")
	keyFile := fs.String("key-file", filepath.Join(dir, "peer.key"), "private peer token file")
	listen := fs.String("listen", "127.0.0.1:9774", "peer listener (TLS required for LAN)")
	certificate := fs.String("tls-cert", "", "PEM TLS certificate")
	privateKey := fs.String("tls-key", "", "PEM TLS private key")
	peerURL := fs.String("peer", "", "remote HTTPS agent peer URL")
	grant := fs.String("grant", "", "approved source grant ID")
	relative := fs.String("relative", "", "relative source file within approved grant")
	sourceID := fs.String("source-id", "", "remote agent identity")
	staging := fs.String("staging", filepath.Join(dir, "incoming"), "private destination staging directory")
	destinationGrant := fs.String("destination-grant", "", "locally approved writable destination grant")
	transferID := fs.String("transfer", "", "verified staged transfer ID")
	if err = fs.Parse(args[1:]); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return errors.New("unexpected peer arguments")
	}
	switch args[0] {
	case "publish":
		if *destinationGrant == "" || *transferID == "" {
			return errors.New("peer publish requires -destination-grant and -transfer")
		}
		if _, err := agentclient.Read(*identity); err != nil {
			return fmt.Errorf("local agent not enrolled: %w", err)
		}
		permissions, err := agentpolicy.Open(*grantsPath)
		if err != nil {
			return err
		}
		stage, err := agenttransfer.Open(*staging, agenttransfer.DefaultMaxPendingBytes)
		if err != nil {
			return err
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		result, err := (agentpublish.Publisher{Grants: permissions, Transfers: stage}).Publish(ctx, *destinationGrant, *transferID)
		if err != nil {
			return err
		}
		fmt.Fprintf(output, "Published verified file %s (%d bytes, SHA-256 %s) into approved destination %s. Existing files were not replaced.\n", result.Path, result.Bytes, result.SHA256, result.GrantID)
		return nil
	case "init":
		key, err := agentpeer.GenerateKey(*keyFile)
		if err != nil {
			return err
		}
		fmt.Fprintln(output, "Device peer access token (displayed once; store securely):")
		fmt.Fprintln(output, key)
		fmt.Fprintln(output, "Only share this key with a trusted synchronization controller. Keep this terminal output private.")
		return nil
	case "inventory":
		if *peerURL == "" || *grant == "" {
			return errors.New("peer inventory requires -peer and -grant")
		}
		fmt.Fprintln(output, "Enter the source device peer access token:")
		token, err := readPeerToken(input)
		if err != nil {
			return err
		}
		entries, err := (agentpeer.Client{Server: *peerURL, Token: token}).Inventory(context.Background(), *grant)
		if err != nil {
			return err
		}
		for _, e := range entries {
			fmt.Fprintf(output, "%s (%d bytes)\n", e.Path, e.Size)
		}
		fmt.Fprintf(output, "Verified inventory response: %d files (no files modified).\n", len(entries))
		return nil
	case "stage", "receive":
		if args[0] == "receive" && *destinationGrant == "" {
			return errors.New("peer receive requires -destination-grant")
		}
		if *peerURL == "" || *grant == "" || *relative == "" || *sourceID == "" {
			return errors.New("peer stage requires -peer -grant -relative and -source-id")
		}
		if _, err := agentclient.Read(*identity); err != nil {
			return fmt.Errorf("local agent not enrolled: %w", err)
		}
		fmt.Fprintln(output, "Enter the source device peer access token:")
		token, err := readPeerToken(input)
		if err != nil {
			return err
		}
		store, err := agenttransfer.Open(*staging, agenttransfer.DefaultMaxPendingBytes)
		if err != nil {
			return err
		}
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		done, err := (agentpeer.Client{Server: *peerURL, Token: token}).Stage(ctx, *grant, *relative, *sourceID, store)
		if err != nil {
			return err
		}
		fmt.Fprintf(output, "Verified %s: %d bytes staged under %s. No destination files overwritten.\n", done.RelativePath, done.ReceivedBytes, done.ID)
		if args[0] == "receive" {
			permissions, err := agentpolicy.Open(*grantsPath)
			if err != nil {
				return err
			}
			result, err := (agentpublish.Publisher{Grants: permissions, Transfers: store}).Publish(ctx, *destinationGrant, done.ID)
			if err != nil {
				return fmt.Errorf("transfer verified and retained in staging as %s; destination publish refused: %w", done.ID, err)
			}
			fmt.Fprintf(output, "Received %s: %d bytes published to approved destination %s (SHA-256: %s).\n", result.Path, result.Bytes, result.GrantID, result.SHA256)
		}
		return nil
	case "serve":
		credentials, err := agentclient.Read(*identity)
		if err != nil {
			return fmt.Errorf("enroll the agent before serving files: %w", err)
		}
		store, err := agentpolicy.Open(*grantsPath)
		if err != nil {
			return err
		}
		token, err := agentpeer.ReadKey(*keyFile)
		if err != nil {
			return fmt.Errorf("create a peer token with 'peer init' first: %w", err)
		}
		handler, err := agentpeer.NewHandler(store, token)
		if err != nil {
			return err
		}
		host, _, err := net.SplitHostPort(*listen)
		if err != nil {
			return err
		}
		hostIP := net.ParseIP(host)
		local := strings.EqualFold(host, "localhost") || hostIP != nil && hostIP.IsLoopback()
		if (*certificate == "") != (*privateKey == "") {
			return errors.New("provide both -tls-cert and -tls-key")
		}
		if !local && *certificate == "" {
			return errors.New("non-loopback agent file service requires TLS certificate and key")
		}
		// The peer endpoint is usable only with a recent successful heartbeat to
		// the enrolled controller. Losing authentication fails closed within 90s.
		var lastAccepted atomic.Int64
		handler.Active = func() bool {
			last := lastAccepted.Load()
			return last != 0 && time.Since(time.Unix(0, last)) <= 90*time.Second
		}
		agent := agentclient.Client{}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		heartbeat := func() error {
			check, done := context.WithTimeout(ctx, 20*time.Second)
			defer done()
			err := agent.Heartbeat(check, credentials, runtime.GOOS, runtime.GOARCH, version)
			if err == nil {
				lastAccepted.Store(time.Now().UnixNano())
			}
			return err
		}
		if err = heartbeat(); err != nil {
			return fmt.Errorf("controller must accept heartbeat before serving: %w", err)
		}
		server := &http.Server{Addr: *listen, Handler: handler, ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10}
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if err := heartbeat(); err != nil {
						if errors.Is(err, agentclient.ErrRevoked) {
							lastAccepted.Store(0)
							stop()
							return
						}
						log.Printf("peer controller heartbeat unavailable (access expires within 90 seconds): %v", err)
					}
				}
			}
		}()
		go func() {
			<-ctx.Done()
			shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = server.Shutdown(shutdown)
		}()
		fmt.Fprintf(output, "Syncria peer serving approved read-only folders on %s; central authorization required.\n", *listen)
		if *certificate != "" {
			err = server.ListenAndServeTLS(*certificate, *privateKey)
		} else {
			err = server.ListenAndServe()
		}
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	default:
		return fmt.Errorf("unknown peer command %q; choose init, serve, inventory, stage, receive, or publish", args[0])
	}
}
func readPeerToken(input io.Reader) (string, error) {
	if input == nil {
		return "", errors.New("token input required")
	}
	// Read a whole line, with an upper bound. A single Reader.Read can return
	// a partial token when stdin is a pipe or pseudo-terminal.
	line, err := bufio.NewReader(io.LimitReader(input, 128)).ReadString('\n')
	if err != nil && !(errors.Is(err, io.EOF) && line != "") {
		return "", err
	}
	token := strings.TrimSpace(line)
	if len(token) != 43 {
		return "", errors.New("invalid peer token input")
	}
	return token, nil
}
