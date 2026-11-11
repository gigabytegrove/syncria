// Package agentclient provides Syncria's optional, identity-only device client.
// It cannot enumerate, read, transfer, or modify local files.
package agentclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const UserAgent = "Syncria-Agent/1.0.0-dev"

var ErrRevoked = errors.New("agent credential rejected or revoked")

type Credentials struct {
	Server     string    `json:"server"`
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Secret     string    `json:"secret"`
	EnrolledAt time.Time `json:"enrolled_at"`
}

type Identity struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	EnrolledAt time.Time `json:"enrolled_at"`
}

type Client struct{ HTTP *http.Client }

func ValidServer(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || u == nil || u.User != nil || u.Host == "" || u.RawQuery != "" || u.Fragment != "" || u.Path != "" && u.Path != "/" {
		return "", errors.New("invalid Syncria server URL")
	}
	if u.Scheme != "https" {
		if u.Scheme != "http" {
			return "", errors.New("Syncria agent connection must use HTTPS")
		}
		host := u.Hostname()
		ip := net.ParseIP(host)
		if !strings.EqualFold(host, "localhost") && (ip == nil || !ip.IsLoopback()) {
			return "", errors.New("plaintext HTTP allowed only for loopback development")
		}
	}
	return strings.TrimRight(u.String(), "/"), nil
}
func (c Client) client() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: 20 * time.Second, Transport: &http.Transport{Proxy: http.ProxyFromEnvironment, TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12}}}
}
func (c Client) post(ctx context.Context, server, route string, body any, id, secret string, out any) error {
	base, err := ValidServer(server)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", base+route, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", UserAgent)
	if id != "" {
		req.Header.Set("X-Syncria-Agent-ID", id)
		req.Header.Set("Authorization", "Bearer "+secret)
	}
	hc := c.client()
	// Never follow a redirect with a one-use pairing code or bearer credential.
	copyClient := *hc
	copyClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := copyClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			return ErrRevoked
		}
		return fmt.Errorf("Syncria agent API returned HTTP %d", resp.StatusCode)
	}
	if err = json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(out); err != nil {
		return fmt.Errorf("decode Syncria agent response: %w", err)
	}
	return nil
}
func (c Client) Enroll(ctx context.Context, server, token, platform, arch, version string) (Credentials, error) {
	if len(token) != 43 {
		return Credentials{}, errors.New("invalid pairing token")
	}
	var result struct {
		Device     Identity `json:"device"`
		Credential string   `json:"credential"`
	}
	err := c.post(ctx, server, "/api/v1/agents/enroll", map[string]string{"token": token, "platform": platform, "arch": arch, "version": version}, "", "", &result)
	if err != nil {
		return Credentials{}, err
	}
	if result.Device.ID == "" || len(result.Credential) != 43 {
		return Credentials{}, errors.New("Syncria returned incomplete enrollment credentials")
	}
	base, _ := ValidServer(server)
	return Credentials{Server: base, ID: result.Device.ID, Name: result.Device.Name, Secret: result.Credential, EnrolledAt: result.Device.EnrolledAt}, nil
}
func (c Client) Heartbeat(ctx context.Context, identity Credentials, platform, arch, version string) error {
	if identity.ID == "" || len(identity.Secret) != 43 {
		return errors.New("agent is not enrolled")
	}
	var result struct {
		Device Identity `json:"device"`
	}
	err := c.post(ctx, identity.Server, "/api/v1/agents/heartbeat", map[string]string{"platform": platform, "arch": arch, "version": version}, identity.ID, identity.Secret, &result)
	if err != nil {
		return err
	}
	if result.Device.ID != identity.ID {
		return errors.New("Syncria returned an unexpected device identity")
	}
	return nil
}
func DefaultFile() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "syncria-agent", "identity.json"), nil
}
func Read(path string) (Credentials, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return Credentials{}, err
	}
	if !fi.Mode().IsRegular() {
		return Credentials{}, errors.New("agent identity must be a regular file")
	}
	if runtime.GOOS != "windows" && fi.Mode().Perm()&0077 != 0 {
		return Credentials{}, errors.New("agent identity file permissions are too broad")
	}
	f, err := os.Open(path)
	if err != nil {
		return Credentials{}, err
	}
	defer f.Close()
	var identity Credentials
	decoder := json.NewDecoder(io.LimitReader(f, 8192))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&identity); err != nil {
		return Credentials{}, err
	}
	if identity.ID == "" || len(identity.Secret) != 43 {
		return Credentials{}, errors.New("invalid saved agent credentials")
	}
	if _, err = ValidServer(identity.Server); err != nil {
		return Credentials{}, err
	}
	return identity, nil
}
func Save(path string, identity Credentials) error {
	if identity.ID == "" || len(identity.Secret) != 43 {
		return errors.New("invalid agent identity")
	}
	if _, err := ValidServer(identity.Server); err != nil {
		return err
	}
	if !filepath.IsAbs(path) {
		return errors.New("absolute agent identity file path required")
	}
	if fi, err := os.Lstat(path); err == nil {
		if !fi.Mode().IsRegular() {
			return errors.New("agent identity path is not a regular file")
		}
		return errors.New("agent already enrolled; refusing to replace saved credentials")
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(identity, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".agent-identity-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = f.Chmod(0600); err != nil {
		f.Close()
		return err
	}
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	// Link fails if destination already exists, avoiding accidental replacement
	// and preserving the original device identity on second enrollment.
	if err = os.Link(f.Name(), path); err != nil {
		return fmt.Errorf("install agent identity: %w", err)
	}
	return nil
}
