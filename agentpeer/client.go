package agentpeer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gigabytegrove/syncria/agentfiles"
	"github.com/gigabytegrove/syncria/agenttransfer"
)

type Client struct {
	Server string
	Token  string
	HTTP   *http.Client
}

func (c Client) prepare(route string) (string, error) {
	u, err := url.Parse(c.Server)
	if err != nil || u == nil || u.User != nil || u.Host == "" || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return "", errors.New("invalid peer URL")
	}
	if u.Scheme != "https" {
		ip := net.ParseIP(u.Hostname())
		if u.Scheme != "http" || (!strings.EqualFold(u.Hostname(), "localhost") && (ip == nil || !ip.IsLoopback())) {
			return "", errors.New("peer connection requires HTTPS")
		}
	}
	if len(c.Token) != 43 {
		return "", errors.New("peer token not configured")
	}
	return strings.TrimRight(c.Server, "/") + route, nil
}
func (c Client) request(ctx context.Context, method, route string, headers http.Header) (*http.Response, error) {
	endpoint, err := c.prepare(route)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/json")
	for key, values := range headers {
		for _, v := range values {
			req.Header.Add(key, v)
		}
	}
	client := c.HTTP
	if client == nil {
		client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyFromEnvironment}, Timeout: 0}
	}
	clone := *client
	clone.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return clone.Do(req)
}
func (c Client) Inventory(ctx context.Context, id string) ([]agentfiles.Entry, error) {
	if !validGrantID(id) { return nil, errors.New("invalid approved source grant") }
	resp, err := c.request(ctx, "GET", "/v1/roots/"+url.PathEscape(id)+"/inventory", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("remote inventory HTTP %d", resp.StatusCode)
	}
	// An incomplete or truncated inventory is not an empty source.
	payload, err := io.ReadAll(io.LimitReader(resp.Body, MaxInventoryBytes+1))
	if err != nil { return nil, fmt.Errorf("read source inventory: %w", err) }
	if len(payload) > MaxInventoryBytes { return nil, errors.New("remote inventory exceeds byte limit") }
	var result struct { Entries *[]agentfiles.Entry `json:"entries"` }
	dec := json.NewDecoder(bytes.NewReader(payload))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&result); err != nil { return nil, fmt.Errorf("invalid remote inventory: %w", err) }
	var trailing any
	if err := dec.Decode(&trailing); !errors.Is(err, io.EOF) { return nil, errors.New("remote inventory contains trailing data") }
	if result.Entries == nil { return nil, errors.New("remote inventory missing entries") }
	if len(*result.Entries) > agentfiles.MaxEntries { return nil, errors.New("remote inventory too large") }
	return *result.Entries, nil
}

// Stage downloads a selected source file through the authenticated peer
// service, appending individually hashed chunks to durable local staging.
// A retry resumes from the committed byte offset. The destination is not
// overwritten, and synchronization conflict/deletion policy stays external.
func (c Client) Stage(ctx context.Context, grantID, relative, sourceID string, store *agenttransfer.Store) (agenttransfer.Transfer, error) {
	if store == nil {
		return agenttransfer.Transfer{}, errors.New("transfer store required")
	}
	if !validGrantID(grantID) || !validPeerRelative(relative) || sourceID == "" || len(sourceID) > 128 { return agenttransfer.Transfer{}, errors.New("invalid approved transfer source") }
	// Escape each component: reserved URL syntax may be part of a filename.
	route := "/v1/roots/" + url.PathEscape(grantID) + "/files/" + peerEncodedPath(relative)
	head, err := c.request(ctx, "HEAD", route, nil)
	if err != nil {
		return agenttransfer.Transfer{}, err
	}
	_ = head.Body.Close()
	if head.StatusCode != 200 {
		return agenttransfer.Transfer{}, fmt.Errorf("remote HEAD HTTP %d", head.StatusCode)
	}
	size, err := strconv.ParseInt(head.Header.Get("Content-Length"), 10, 64)
	if err != nil || size < 0 {
		return agenttransfer.Transfer{}, errors.New("source size unavailable")
	}
	digest := head.Header.Get("X-Syncria-SHA256")
	if len(digest) != 64 {
		return agenttransfer.Transfer{}, errors.New("remote checksum unavailable")
	}
	if _, err = hex.DecodeString(digest); err != nil {
		return agenttransfer.Transfer{}, err
	}
	request := agenttransfer.Request{SourceID: sourceID, Revision: digest, RelativePath: relative, TotalBytes: size, SHA256: digest}
	tr, err := store.Begin(request)
	if err != nil {
		return agenttransfer.Transfer{}, err
	}
	if tr.Status == agenttransfer.StatusComplete {
		return tr, nil
	}
	if size == 0 {
		return store.Complete(tr.ID)
	}
	headers := make(http.Header)
	headers.Set("If-Match", `"`+digest+`"`)
	if tr.ReceivedBytes > 0 {
		headers.Set("Range", fmt.Sprintf("bytes=%d-", tr.ReceivedBytes))
	}
	resp, err := c.request(ctx, "GET", route, headers)
	if err != nil {
		return tr, err
	}
	defer resp.Body.Close()
	expectedStatus := 200
	if tr.ReceivedBytes > 0 {
		expectedStatus = http.StatusPartialContent
	}
	if resp.StatusCode != expectedStatus {
		return tr, fmt.Errorf("remote GET HTTP %d, expected %d", resp.StatusCode, expectedStatus)
	}
	if tr.ReceivedBytes > 0 {
		want := fmt.Sprintf("bytes %d-%d/%d", tr.ReceivedBytes, size-1, size)
		if resp.Header.Get("Content-Range") != want {
			return tr, errors.New("remote range does not match committed checkpoint")
		}
	}
	if resp.Header.Get("ETag") != `"`+digest+`"` {
		return tr, errors.New("remote ETag changed")
	}
	buf := make([]byte, agenttransfer.MaxChunkBytes)
	offset := tr.ReceivedBytes
	offsetStart := offset
	for offset < size {
		if err = ctx.Err(); err != nil {
			return tr, err
		}
		remaining := size - offset
		n := len(buf)
		if remaining < int64(n) {
			n = int(remaining)
		}
		_, err = io.ReadFull(resp.Body, buf[:n])
		if err != nil {
			return tr, err
		}
		sum := sha256.Sum256(buf[:n])
		tr, err = store.Append(tr.ID, offset, bytes.NewReader(buf[:n]), n, hex.EncodeToString(sum[:]))
		if err != nil {
			return tr, err
		}
		offset += int64(n)
	}
	var excess [1]byte
	if n, e := resp.Body.Read(excess[:]); n > 0 || e == nil {
		return tr, errors.New("remote sent unexpected trailing bytes")
	}
	if resp.ContentLength >= 0 && resp.ContentLength != size-offsetStart {
		return tr, errors.New("remote content length mismatch")
	}
	return store.Complete(tr.ID)
}

// CheckHealth is intentionally lightweight and bounded for use in operator
// connectivity checks; it is not evidence that sync jobs have completed.
func (c Client) CheckHealth(ctx context.Context, grantID string) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	_, err := c.Inventory(ctx, grantID)
	return err
}
