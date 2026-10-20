package agentpeer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// OpenVerified opens a remote approved source without caching its payload on
// the controller. Content is SHA-256 verified while it is consumed. A digest,
// length or truncated-stream failure must be observed before the destination
// commits its temporary file.
func (c Client) OpenVerified(ctx context.Context, grant, relative string) (io.ReadCloser, int64, error) {
	if !validGrantID(grant) || !validPeerRelative(relative) {
		return nil, 0, errors.New("invalid approved source selection")
	}
	route := "/v1/roots/" + url.PathEscape(grant) + "/files/" + peerEncodedPath(relative)
	head, err := c.request(ctx, http.MethodHead, route, nil)
	if err != nil {
		return nil, 0, err
	}
	_ = head.Body.Close()
	if head.StatusCode != 200 {
		return nil, 0, fmt.Errorf("remote file HEAD HTTP %d", head.StatusCode)
	}
	size, err := strconv.ParseInt(head.Header.Get("Content-Length"), 10, 64)
	if err != nil || size < 0 || size > 1<<40 {
		return nil, 0, errors.New("invalid remote file size")
	}
	digest := head.Header.Get("X-Syncria-SHA256")
	if len(digest) != 64 || strings.ToLower(digest) != digest {
		return nil, 0, errors.New("missing source checksum")
	}
	decoded, err := hex.DecodeString(digest)
	if err != nil || len(decoded) != sha256.Size {
		return nil, 0, errors.New("invalid source checksum")
	}
	headers := http.Header{}
	headers.Set("If-Match", `"`+digest+`"`)
	reply, err := c.request(ctx, http.MethodGet, route, headers)
	if err != nil {
		return nil, 0, err
	}
	if reply.StatusCode != 200 {
		reply.Body.Close()
		return nil, 0, fmt.Errorf("remote file GET HTTP %d", reply.StatusCode)
	}
	if reply.Header.Get("ETag") != `"`+digest+`"` || reply.ContentLength != size {
		reply.Body.Close()
		return nil, 0, errors.New("source changed between metadata and stream")
	}
	return &verifiedBody{body: reply.Body, expected: decoded, hash: sha256.New(), size: size}, size, nil
}
func validPeerRelative(relative string) bool {
	if relative == "" || len(relative) > 2048 || strings.HasPrefix(relative, "/") || strings.ContainsAny(relative, "\\\x00\r\n") {
		return false
	}
	for _, part := range strings.Split(relative, "/") {
		if part == "" || part == "." || part == ".." || len(part) > 255 || strings.HasPrefix(part, ".syncria-") {
			return false
		}
		for _, ch := range part {
			if ch < 32 || ch == 127 {
				return false
			}
		}
	}
	return true
}

type verifiedBody struct {
	body           io.ReadCloser
	hash           hash.Hash
	expected       []byte
	size, received int64
	finished       bool
}

func (v *verifiedBody) Read(p []byte) (int, error) {
	if v.finished {
		return 0, io.EOF
	}
	if v.received == v.size {
		var probe [1]byte
		n, err := v.body.Read(probe[:])
		if n != 0 || err == nil {
			return 0, errors.New("source returned unexpected trailing bytes")
		}
		if err != io.EOF {
			return 0, err
		}
		v.finished = true
		if !equalDigest(v.hash.Sum(nil), v.expected) {
			return 0, errors.New("source checksum changed during transfer")
		}
		return 0, io.EOF
	}
	max := int64(len(p))
	if remaining := v.size - v.received; max > remaining {
		max = remaining
	}
	if max == 0 {
		return 0, nil
	}
	n, err := v.body.Read(p[:int(max)])
	if n > 0 {
		v.received += int64(n)
		_, _ = v.hash.Write(p[:n])
	}
	if err == io.EOF {
		if v.received == v.size {
			v.finished = true
			if !equalDigest(v.hash.Sum(nil), v.expected) {
				return n, errors.New("source checksum changed during transfer")
			}
			return n, io.EOF
		}
		return n, io.ErrUnexpectedEOF
	}
	return n, err
}
func (v *verifiedBody) Close() error { return v.body.Close() }
func equalDigest(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := range a {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}

// Encode every segment separately, preserving folder separators while
// preventing '?' and '#' in legitimate file names from becoming URL syntax.
func peerEncodedPath(relative string) string {
	parts := strings.Split(relative, "/")
	for i, component := range parts {
		parts[i] = url.PathEscape(component)
	}
	return strings.Join(parts, "/")
}

 // validGrantID accepts only the lowercase hexadecimal IDs created by
 // locally approved folder grants, never arbitrary remote URI segments.
 func validGrantID(id string) bool {
  if len(id)!=32 {return false}
  for _,ch:=range id {
   if !((ch>='0'&&ch<='9')||(ch>='a'&&ch<='f')) {return false}
  }
  return true
 }
