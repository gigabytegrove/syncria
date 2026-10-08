package main

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var appVersion = "0.1.0-alpha"
const folderMime = "application/vnd.google-apps.folder"
const scope = "https://www.googleapis.com/auth/drive"

type Account struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Label        string    `json:"label"`
	RefreshToken string    `json:"refresh_token"`
	AccessToken  string    `json:"access_token"`
	Expires      time.Time `json:"expires"`
}
type Item struct {
	ID       string `json:"id"`
	MD5      string `json:"md5"`
	LocalMD5 string `json:"local_md5"`
	Folder   bool   `json:"folder"`
}
type Job struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	AccountID   string          `json:"account_id"`
	RemoteID    string          `json:"remote_id"`
	Local       string          `json:"local"`
	Interval    int             `json:"interval"`
	Enabled     bool            `json:"enabled"`
	State       map[string]Item `json:"state"`
	LastRun     time.Time       `json:"last_run"`
	LastError   string          `json:"last_error"`
	LastSummary string          `json:"last_summary"`
}
type Config struct {
	Salt         string    `json:"salt"`
	PasswordHash string    `json:"password_hash"`
	Secret       string    `json:"secret"`
	OAuthID      string    `json:"oauth_id"`
	OAuthSecret  string    `json:"oauth_secret"`
	BaseURL      string    `json:"base_url"`
	Accounts     []Account `json:"accounts"`
	Jobs         []Job     `json:"jobs"`
	Shares       []StorageShare `json:"shares,omitempty"`
}
type App struct {
	mu     sync.Mutex
	runs   sync.Map
	cfg    Config
	file   string
	client *http.Client
	oauth  map[string]time.Time
}
type driveFile struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	MimeType    string   `json:"mimeType"`
	MD5Checksum string   `json:"md5Checksum"`
	Parents     []string `json:"parents"`
	Size        string   `json:"size"`
}
type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int    `json:"expires_in"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
func b64rand() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
func hashPass(p, s string) string {
	salt, _ := hex.DecodeString(s)
	mac := func(key, data []byte) []byte { h := hmac.New(sha256.New, key); h.Write(data); return h.Sum(nil) }
	key := []byte(p)
	block := []byte{0, 0, 0, 1}
	u := mac(key, append(append([]byte{}, salt...), block...))
	out := append([]byte{}, u...)
	for i := 1; i < 210000; i++ {
		u = mac(key, u)
		for j := range out {
			out[j] ^= u[j]
		}
	}
	return hex.EncodeToString(out)
}
func passwordOK(p string, c Config) bool {
	if c.PasswordHash == "" {
		return false
	}
	got, _ := hex.DecodeString(hashPass(p, c.Salt))
	expected, _ := hex.DecodeString(c.PasswordHash)
	return subtle.ConstantTimeCompare(got, expected) == 1
}
func (a *App) save() error {
	b, err := json.MarshalIndent(a.cfg, "", "  ")
	if err != nil {
		return err
	}
	tmp := a.file + ".tmp"
	if err = os.WriteFile(tmp, b, 0600); err != nil {
		return err
	}
	if err = os.Rename(tmp, a.file); err != nil {
		return err
	}
	return nil
}
func (a *App) account(id string) *Account {
	for i := range a.cfg.Accounts {
		if a.cfg.Accounts[i].ID == id {
			return &a.cfg.Accounts[i]
		}
	}
	return nil
}
func (a *App) job(id string) *Job {
	for i := range a.cfg.Jobs {
		if a.cfg.Jobs[i].ID == id {
			return &a.cfg.Jobs[i]
		}
	}
	return nil
}
func (a *App) token(ctx context.Context, id string) (string, error) {
	a.mu.Lock()
	ac := a.account(id)
	if ac == nil {
		a.mu.Unlock()
		return "", errors.New("account not found")
	}
	if ac.AccessToken != "" && time.Until(ac.Expires) > 2*time.Minute {
		v := ac.AccessToken
		a.mu.Unlock()
		return v, nil
	}
	refresh := ac.RefreshToken
	cid := a.cfg.OAuthID
	secret := a.cfg.OAuthSecret
	a.mu.Unlock()
	v := url.Values{"client_id": {cid}, "client_secret": {secret}, "refresh_token": {refresh}, "grant_type": {"refresh_token"}}
	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", strings.NewReader(v.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var t tokenResponse
	if err = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&t); err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("token refresh: %s: %s", t.Error, t.ErrorDescription)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if ac = a.account(id); ac != nil {
		ac.AccessToken = t.AccessToken
		ac.Expires = time.Now().Add(time.Duration(t.ExpiresIn) * time.Second)
		_ = a.save()
	}
	return t.AccessToken, nil
}
func (a *App) drive(ctx context.Context, account, method, endpoint string, body io.Reader, ctype string) (*http.Response, error) {
	tok, err := a.token(ctx, account)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	if ctype != "" {
		req.Header.Set("Content-Type", ctype)
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
		resp.Body.Close()
		return nil, fmt.Errorf("drive HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	return resp, nil
}
func (a *App) apiJSON(ctx context.Context, acc, method, endpoint string, payload any, out any) error {
	var r io.Reader
	ctype := ""
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		r = strings.NewReader(string(b))
		ctype = "application/json"
	}
	resp, err := a.drive(ctx, acc, method, endpoint, r, ctype)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if out == nil {
		io.Copy(io.Discard, resp.Body)
		return nil
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(out)
}
func (a *App) folderChildren(ctx context.Context, acc, parent string) ([]driveFile, error) {
	var items []driveFile
	page := ""
	for {
		q := url.Values{}
		q.Set("q", fmt.Sprintf("'%s' in parents and trashed = false", strings.ReplaceAll(parent, "'", "\\'")))
		q.Set("fields", "nextPageToken,files(id,name,mimeType,md5Checksum,parents,size)")
		q.Set("pageSize", "1000")
		if page != "" {
			q.Set("pageToken", page)
		}
		var r struct {
			NextPageToken string      `json:"nextPageToken"`
			Files         []driveFile `json:"files"`
		}
		if err := a.apiJSON(ctx, acc, "GET", "https://www.googleapis.com/drive/v3/files?"+q.Encode(), nil, &r); err != nil {
			return nil, err
		}
		items = append(items, r.Files...)
		if r.NextPageToken == "" {
			break
		}
		page = r.NextPageToken
	}
	return items, nil
}
func safeName(n string) bool {
	return n != "" && n != "." && n != ".." && !strings.ContainsAny(n, "/\\\x00") && n != ".sync-agent-trash" && n != ".DS_Store"
}
func (a *App) remoteTree(ctx context.Context, acc, root string) (map[string]driveFile, error) {
	res := map[string]driveFile{}
	seen := map[string]bool{}
	var walk func(string, string, int) error
	walk = func(id, p string, depth int) error {
		if depth > 64 {
			return errors.New("remote folder depth exceeded")
		}
		if seen[id] {
			return errors.New("remote folder cycle")
		}
		seen[id] = true
		children, err := a.folderChildren(ctx, acc, id)
		if err != nil {
			return err
		}
		names := map[string]bool{}
		for _, f := range children {
			if !safeName(f.Name) {
				return fmt.Errorf("unsafe remote filename: %q", f.Name)
			}
			if names[strings.ToLower(f.Name)] {
				return fmt.Errorf("duplicate/case-colliding remote name in %s: %s", p, f.Name)
			}
			names[strings.ToLower(f.Name)] = true
			path := f.Name
			if p != "" {
				path = p + "/" + f.Name
			}
			if f.MimeType != folderMime && strings.HasPrefix(f.MimeType, "application/vnd.google-apps.") {
				return fmt.Errorf("Google-native document %q requires export; sync paused to protect data", path)
			}
			res[path] = f
			if f.MimeType == folderMime {
				if err := walk(f.ID, path, depth+1); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return res, walk(root, "", 0)
}
func md5File(path string) (string, error) {
	f, e := os.Open(path)
	if e != nil {
		return "", e
	}
	defer f.Close()
	h := md5.New()
	if _, e = io.Copy(h, f); e != nil {
		return "", e
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func localTree(root string) (map[string]Item, error) {
	out := map[string]Item{}
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if p == root {
			return nil
		}
		rel, e := filepath.Rel(root, p)
		if e != nil {
			return e
		}
		rel = filepath.ToSlash(rel)
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink unsupported: %s", rel)
		}
		if !d.IsDir() && !d.Type().IsRegular() {
			return fmt.Errorf("special file unsupported: %s", rel)
		}
		if !safeName(d.Name()) {
			return fmt.Errorf("unsafe local filename: %s", rel)
		}
		item := Item{Folder: d.IsDir()}
		if !item.Folder {
			item.LocalMD5, e = md5File(p)
			if e != nil {
				return e
			}
		}
		out[rel] = item
		return nil
	})
	return out, err
}
func localPath(root, rel string) (string, error) {
	if rel == "" {
		return "", errors.New("empty relative path")
	}
	parts := strings.Split(rel, "/")
	for _, v := range parts {
		if !safeName(v) {
			return "", errors.New("unsafe local path")
		}
	}
	return filepath.Join(append([]string{root}, parts...)...), nil
}
func parentRel(rel string) string {
	p := strings.LastIndex(rel, "/")
	if p < 0 {
		return ""
	}
	return rel[:p]
}
func basename(rel string) string { p := strings.LastIndex(rel, "/"); return rel[p+1:] }
func (a *App) makeFolder(ctx context.Context, acc, parent, name string) (driveFile, error) {
	var f driveFile
	err := a.apiJSON(ctx, acc, "POST", "https://www.googleapis.com/drive/v3/files?fields=id,name,mimeType", map[string]any{"name": name, "mimeType": folderMime, "parents": []string{parent}}, &f)
	return f, err
}
func (a *App) upload(ctx context.Context, acc, path, name, parent, id string) (driveFile, error) {
	f, e := os.Open(path)
	if e != nil {
		return driveFile{}, e
	}
	defer f.Close()
	stat, e := f.Stat()
	if e != nil {
		return driveFile{}, e
	}
	ct := mime.TypeByExtension(filepath.Ext(path))
	if ct == "" {
		ct = "application/octet-stream"
	}
	meta := map[string]any{"name": name}
	method := "POST"
	target := "https://www.googleapis.com/upload/drive/v3/files?uploadType=resumable&fields=id,name,mimeType,md5Checksum"
	if id != "" {
		method = "PATCH"
		target = "https://www.googleapis.com/upload/drive/v3/files/" + url.PathEscape(id) + "?uploadType=resumable&fields=id,name,mimeType,md5Checksum"
	} else {
		meta["parents"] = []string{parent}
	}
	data, _ := json.Marshal(meta)
	tok, e := a.token(ctx, acc)
	if e != nil {
		return driveFile{}, e
	}
	req, e := http.NewRequestWithContext(ctx, method, target, strings.NewReader(string(data)))
	if e != nil {
		return driveFile{}, e
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Upload-Content-Type", ct)
	req.Header.Set("X-Upload-Content-Length", strconv.FormatInt(stat.Size(), 10))
	resp, e := a.client.Do(req)
	if e != nil {
		return driveFile{}, e
	}
	resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return driveFile{}, fmt.Errorf("start upload HTTP %d", resp.StatusCode)
	}
	session := resp.Header.Get("Location")
	if session == "" {
		return driveFile{}, errors.New("no resumable upload location")
	}
	if !strings.HasPrefix(session, "https://www.googleapis.com/") && !strings.HasPrefix(session, "https://www.google.com/") {
		return driveFile{}, errors.New("unexpected upload destination")
	}
	req, e = http.NewRequestWithContext(ctx, "PUT", session, f)
	if e != nil {
		return driveFile{}, e
	}
	req.ContentLength = stat.Size()
	req.Header.Set("Content-Type", ct)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, e = a.client.Do(req)
	if e != nil {
		return driveFile{}, e
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
		return driveFile{}, fmt.Errorf("upload HTTP %d: %s", resp.StatusCode, b)
	}
	var result driveFile
	e = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result)
	return result, e
}
func (a *App) download(ctx context.Context, acc, id, path string) error {
	resp, e := a.drive(ctx, acc, "GET", "https://www.googleapis.com/drive/v3/files/"+url.PathEscape(id)+"?alt=media", nil, "")
	if e != nil {
		return e
	}
	defer resp.Body.Close()
	if e = os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		return e
	}
	tmp, e := os.CreateTemp(filepath.Dir(path), ".sync-download-*")
	if e != nil {
		return e
	}
	name := tmp.Name()
	defer os.Remove(name)
	if _, e = io.Copy(tmp, resp.Body); e != nil {
		tmp.Close()
		return e
	}
	if e = tmp.Sync(); e != nil {
		tmp.Close()
		return e
	}
	if e = tmp.Close(); e != nil {
		return e
	}
	return os.Rename(name, path)
}
func (a *App) trashRemote(ctx context.Context, acc, id string) error {
	return a.apiJSON(ctx, acc, "PATCH", "https://www.googleapis.com/drive/v3/files/"+url.PathEscape(id), map[string]any{"trashed": true}, nil)
}
func (a *App) trashLocal(path, jobID, rel string) error {
	dir := filepath.Join(filepath.Dir(a.file), "recovery", jobID, time.Now().UTC().Format("20060102-150405.000000000"))
	target := filepath.Join(dir, filepath.FromSlash(rel))
	if e := os.MkdirAll(filepath.Dir(target), 0700); e != nil {
		return e
	}
	return os.Rename(path, target)
}
func depthKey(s string) int { return strings.Count(s, "/") }
func sortedKeys(m map[string]Item, reverse bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		di, dj := depthKey(keys[i]), depthKey(keys[j])
		if di != dj {
			if reverse {
				return di > dj
			}
			return di < dj
		}
		return keys[i] < keys[j]
	})
	return keys
}
func sameLocal(prev, now Item) bool {
	return prev.Folder == now.Folder && (prev.Folder || prev.LocalMD5 == now.LocalMD5)
}
func sameRemote(prev Item, now driveFile) bool {
	return prev.ID == now.ID && prev.Folder == (now.MimeType == folderMime) && (prev.Folder || prev.MD5 == now.MD5Checksum)
}
func (a *App) syncJob(ctx context.Context, jobID string) error {
	if _, loaded := a.runs.LoadOrStore(jobID, true); loaded {
		return errors.New("job already running")
	}
	defer a.runs.Delete(jobID)
	a.mu.Lock()
	j := a.job(jobID)
	if j == nil {
		a.mu.Unlock()
		return errors.New("job not found")
	}
	job := *j
	job.State = make(map[string]Item, len(j.State))
	for k, v := range j.State {
		job.State[k] = v
	}
	a.mu.Unlock()
	ctx, cancel := context.WithTimeout(ctx, 45*time.Minute)
	defer cancel()
	err := a.execute(ctx, &job)
	a.mu.Lock()
	if actual := a.job(jobID); actual != nil {
		actual.LastRun = time.Now()
		actual.LastError = ""
		if err != nil {
			actual.LastError = err.Error()
		}
		actual.LastSummary = job.LastSummary
		actual.State = job.State
		_ = a.save()
	}
	a.mu.Unlock()
	return err
}
func (a *App) execute(ctx context.Context, j *Job) error {
	root, e := filepath.Abs(j.Local)
	if e != nil {
		return e
	}
	info, e := os.Stat(root)
	if e != nil {
		return fmt.Errorf("local folder not accessible: %w", e)
	}
	if !info.IsDir() {
		return errors.New("local mapping is not a directory")
	}
	remote, e := a.remoteTree(ctx, j.AccountID, j.RemoteID)
	if e != nil {
		return e
	}
	local, e := localTree(root)
	if e != nil {
		return e
	}
	next := map[string]Item{}
	for k, v := range j.State {
		next[k] = v
	}
	count := map[string]int{}
	changed := map[string]bool{}
	all := map[string]Item{}
	for k, v := range local {
		all[k] = v
	}
	for k, v := range remote {
		all[k] = Item{Folder: v.MimeType == folderMime}
	}
	for k, v := range next {
		all[k] = v
	}
	keys := sortedKeys(all, false)
	for _, rel := range keys {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		rv, hasR := remote[rel]
		lv, hasL := local[rel]
		pv, tracked := next[rel]
		isFolder := hasR && rv.MimeType == folderMime || hasL && lv.Folder || tracked && pv.Folder
		if (hasR && rv.MimeType == folderMime) != (hasL && lv.Folder) && hasR && hasL {
			return fmt.Errorf("file/folder type conflict: %s", rel)
		}
		path, e := localPath(root, rel)
		if e != nil {
			return e
		}
		if isFolder {
			switch {
			case hasR && hasL:
				next[rel] = Item{ID: rv.ID, Folder: true}
			case hasR && !hasL:
				if tracked {
					return fmt.Errorf("folder deletion requires manual review to protect descendants: %s", rel)
				}
				if e = os.MkdirAll(path, 0700); e != nil {
					return e
				}
				next[rel] = Item{ID: rv.ID, Folder: true}
				count["download_folders"]++
			case !hasR && hasL:
				if tracked {
					if e = a.trashLocal(path, j.ID, rel); e != nil {
						return e
					}
					delete(next, rel)
					changed[rel] = true
					count["local_recovery"]++
					continue
				}
				parent := j.RemoteID
				if pr := parentRel(rel); pr != "" {
					parent = next[pr].ID
				}
				if parent == "" {
					return fmt.Errorf("missing remote parent for %s", rel)
				}
				f, er := a.makeFolder(ctx, j.AccountID, parent, basename(rel))
				if er != nil {
					return er
				}
				next[rel] = Item{ID: f.ID, Folder: true}
				count["upload_folders"]++
			default:
				delete(next, rel)
			}
			continue
		}
		switch {
		case hasR && hasL:
			if rv.MD5Checksum == "" {
				return fmt.Errorf("no Drive MD5 for %s", rel)
			}
			if lv.LocalMD5 == rv.MD5Checksum {
				next[rel] = Item{ID: rv.ID, MD5: rv.MD5Checksum, LocalMD5: lv.LocalMD5}
				continue
			}
			if !tracked {
				return fmt.Errorf("initial conflict: different files at %s (neither overwritten)", rel)
			}
			localChanged := lv.LocalMD5 != pv.LocalMD5
			remoteChanged := !sameRemote(pv, rv)
			if localChanged && remoteChanged {
				return fmt.Errorf("both sides changed: %s", rel)
			}
			if localChanged {
				f, er := a.upload(ctx, j.AccountID, path, basename(rel), "", rv.ID)
				if er != nil {
					return er
				}
				next[rel] = Item{ID: f.ID, MD5: lv.LocalMD5, LocalMD5: lv.LocalMD5}
				count["uploaded"]++
			} else {
				if e = a.download(ctx, j.AccountID, rv.ID, path); e != nil {
					return e
				}
				next[rel] = Item{ID: rv.ID, MD5: rv.MD5Checksum, LocalMD5: rv.MD5Checksum}
				count["downloaded"]++
			}
		case hasR && !hasL:
			if tracked {
				if !sameRemote(pv, rv) {
					return fmt.Errorf("remote modified after local deletion: %s", rel)
				}
				if e = a.trashRemote(ctx, j.AccountID, rv.ID); e != nil {
					return e
				}
				delete(next, rel)
				count["remote_trash"]++
			} else {
				if e = a.download(ctx, j.AccountID, rv.ID, path); e != nil {
					return e
				}
				next[rel] = Item{ID: rv.ID, MD5: rv.MD5Checksum, LocalMD5: rv.MD5Checksum}
				count["downloaded"]++
			}
		case !hasR && hasL:
			if tracked {
				if !sameLocal(pv, lv) {
					return fmt.Errorf("local changed after remote deletion: %s", rel)
				}
				if e = a.trashLocal(path, j.ID, rel); e != nil {
					return e
				}
				delete(next, rel)
				count["local_recovery"]++
			} else {
				parent := j.RemoteID
				if pr := parentRel(rel); pr != "" {
					parent = next[pr].ID
				}
				if parent == "" {
					return fmt.Errorf("missing remote parent for %s", rel)
				}
				f, er := a.upload(ctx, j.AccountID, path, basename(rel), parent, "")
				if er != nil {
					return er
				}
				next[rel] = Item{ID: f.ID, MD5: lv.LocalMD5, LocalMD5: lv.LocalMD5}
				count["uploaded"]++
			}
		default:
			delete(next, rel)
		}
		j.State = next
		a.mu.Lock()
		if actual := a.job(j.ID); actual != nil {
			actual.State = make(map[string]Item, len(next))
			for k, v := range next {
				actual.State[k] = v
			}
			_ = a.save()
		}
		a.mu.Unlock()
	}
	j.State = next
	parts := []string{}
	for _, k := range []string{"uploaded", "downloaded", "upload_folders", "download_folders", "remote_trash", "local_recovery"} {
		if count[k] > 0 {
			parts = append(parts, fmt.Sprintf("%s: %d", k, count[k]))
		}
	}
	if len(parts) == 0 {
		j.LastSummary = "No changes"
	} else {
		j.LastSummary = strings.Join(parts, ", ")
	}
	return nil
}

func (a *App) authenticated(r *http.Request) bool {
	c, e := r.Cookie("gdsync_session")
	if e != nil {
		return false
	}
	parts := strings.Split(c.Value, ".")
	if len(parts) != 2 {
		return false
	}
	raw, e := base64.RawURLEncoding.DecodeString(parts[0])
	if e != nil {
		return false
	}
	var payload struct {
		Expires int64  `json:"expires"`
		Nonce   string `json:"nonce"`
	}
	if json.Unmarshal(raw, &payload) != nil || time.Now().Unix() > payload.Expires {
		return false
	}
	a.mu.Lock()
	secret := a.cfg.Secret
	a.mu.Unlock()
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[0]))
	want := hex.EncodeToString(mac.Sum(nil))
	return subtle.ConstantTimeCompare([]byte(want), []byte(parts[1])) == 1
}
func (a *App) setSession(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	secret := a.cfg.Secret
	a.mu.Unlock()
	payload, _ := json.Marshal(map[string]any{"expires": time.Now().Add(24 * time.Hour).Unix(), "nonce": newID()})
	part := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(part))
	http.SetCookie(w, &http.Cookie{Name: "gdsync_session", Value: part + "." + hex.EncodeToString(mac.Sum(nil)), HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: r.TLS != nil, Path: "/", MaxAge: 86400})
}
func localPost(r *http.Request) bool {
	if r.Method != "POST" {
		return false
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return r.Header.Get("Sec-Fetch-Site") != "cross-site"
	}
	o, e := url.Parse(origin)
	if e != nil {
		return false
	}
	return strings.EqualFold(o.Host, r.Host) && (o.Scheme == "http" || o.Scheme == "https")
}
func redirect(w http.ResponseWriter, r *http.Request, msg string) {
	http.Redirect(w, r, "/?msg="+url.QueryEscape(msg), http.StatusSeeOther)
}
func (a *App) action(w http.ResponseWriter, r *http.Request) {
	if !localPost(r) {
		http.Error(w, "POST required (same origin)", 405)
		return
	}
	if e := r.ParseForm(); e != nil {
		http.Error(w, "bad form", 400)
		return
	}
	action := r.FormValue("action")
	a.mu.Lock()
	configured := a.cfg.PasswordHash != ""
	a.mu.Unlock()
	if !configured {
		if action != "setup" {
			http.Error(w, "setup required", 403)
			return
		}
		pass := r.FormValue("password")
		if len(pass) < 12 || len(pass) > 1024 {
			redirect(w, r, "Password must be at least 12 characters")
			return
		}
		a.mu.Lock()
		if a.cfg.PasswordHash != "" {
			a.mu.Unlock()
			http.Error(w, "already configured", 409)
			return
		}
		a.cfg.Salt = newID()
		a.cfg.PasswordHash = hashPass(pass, a.cfg.Salt)
		e := a.save()
		a.mu.Unlock()
		if e != nil {
			http.Error(w, e.Error(), 500)
			return
		}
		a.setSession(w, r)
		redirect(w, r, "Welcome to Syncria")
		return
	}
	if action == "login" {
		a.mu.Lock()
		cfg := Config{PasswordHash: a.cfg.PasswordHash, Salt: a.cfg.Salt}
		a.mu.Unlock()
		if !passwordOK(r.FormValue("password"), cfg) {
			time.Sleep(400 * time.Millisecond)
			redirect(w, r, "Invalid password")
			return
		}
		a.setSession(w, r)
		redirect(w, r, "Signed in")
		return
	}
	if !a.authenticated(r) {
		http.Error(w, "Unauthorized", 401)
		return
	}
	switch action {
	case "update_check":
        ctx,cancel:=context.WithTimeout(r.Context(),30*time.Second)
        release,e:=latestRelease(ctx);cancel()
        if e!=nil{redirect(w,r,"Update check failed: "+e.Error());return}
        redirect(w,r,"Latest GitHub release: "+release.Tag+"; running "+appVersion)
        return
    case "update_install", "update_rollback":
        if runtime.GOOS=="windows"{redirect(w,r,"In-app restart not yet supported on Windows; download the release manually");return}
        active:=false
        a.runs.Range(func(k,v any)bool{active=true;return false})
        if active{redirect(w,r,"Update blocked while synchronization is running");return}
        dataDir:=filepath.Dir(a.file)
        if action=="update_install"{
            ctx,cancel:=context.WithTimeout(r.Context(),15*time.Minute)
            rel,e:=latestRelease(ctx)
            if e==nil && strings.TrimPrefix(rel.Tag,"v")==strings.TrimPrefix(appVersion,"v"){cancel();redirect(w,r,"Already on latest release "+rel.Tag);return}
            if e==nil{_,e=updateBinary(ctx,dataDir,rel)}
            cancel()
            if e!=nil{redirect(w,r,"Update failed: "+e.Error());return}
        }else{
            if _,e:=rollbackBinary(dataDir);e!=nil{redirect(w,r,"Rollback failed: "+e.Error());return}
        }
        redirect(w,r,"Update staged. Syncria is restarting; refresh shortly.")
        go func(){time.Sleep(500*time.Millisecond);restartSyncria(dataDir)}()
        return
    case "logout":
		http.SetCookie(w, &http.Cookie{Name: "gdsync_session", MaxAge: -1, Path: "/", HttpOnly: true})
		redirect(w, r, "Signed out")
		return
	case "oauth_settings":
		id := strings.TrimSpace(r.FormValue("client_id"))
		secret := strings.TrimSpace(r.FormValue("client_secret"))
		base := strings.TrimRight(strings.TrimSpace(r.FormValue("base_url")), "/")
		u, e := url.Parse(base)
		if e != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
			redirect(w, r, "Invalid dashboard URL")
			return
		}
		if u.Scheme == "http" && !isLoopbackHost(u.Hostname()) {
			redirect(w, r, "Google OAuth requires HTTPS except for localhost")
			return
		}
		a.mu.Lock()
		if secret == "" {
			secret = a.cfg.OAuthSecret
		}
		a.mu.Unlock()
		if id == "" || secret == "" {
			redirect(w, r, "OAuth client ID and secret required")
			return
		}
		a.mu.Lock()
		a.cfg.OAuthID = id
		a.cfg.OAuthSecret = secret
		a.cfg.BaseURL = base
		e = a.save()
		a.mu.Unlock()
		if e != nil {
			http.Error(w, e.Error(), 500)
			return
		}
		redirect(w, r, "OAuth settings saved")
	case "add_job":
		acc := r.FormValue("account")
		name := strings.TrimSpace(r.FormValue("name"))
		local := strings.TrimSpace(r.FormValue("local"))
		remote := strings.TrimSpace(r.FormValue("remote"))
		if remote == "" || remote == "whole" {
			remote = "root"
		}
		if remote != "root" && !validDriveID(remote) {
			redirect(w, r, "Invalid Drive folder ID")
			return
		}
		if local == "" || name == "" {
			redirect(w, r, "Job name and local path required")
			return
		}
		path, e := filepath.Abs(local)
		if e != nil {
			redirect(w, r, e.Error())
			return
		}
		storageRoot, rootErr := resolvedPath(browseStorageRoot())
		if rootErr != nil {
			redirect(w, r, "Storage root is unavailable")
			return
		}
		realPath, resolveErr := resolvedPath(path)
		if resolveErr != nil || !contained(storageRoot, realPath) {
			redirect(w, r, "Choose a directory from available storage")
			return
		}
		path = realPath
		st, e := os.Stat(path)
		if e != nil || !st.IsDir() {
			redirect(w, r, "Local directory must exist and be accessible")
			return
		}
		interval, e := strconv.Atoi(r.FormValue("interval"))
		if e != nil || interval < 1 || interval > 1440 {
			redirect(w, r, "Interval must be 1 to 1440 minutes")
			return
		}
		a.mu.Lock()
		if a.account(acc) == nil {
			a.mu.Unlock()
			redirect(w, r, "Select an account")
			return
		}
		for _, j := range a.cfg.Jobs {
			if strings.EqualFold(j.Local, path) && j.Enabled {
				a.mu.Unlock()
				redirect(w, r, "Directory is already mapped by an enabled job")
				return
			}
		}
		a.cfg.Jobs = append(a.cfg.Jobs, Job{ID: newID(), Name: name, AccountID: acc, RemoteID: remote, Local: path, Interval: interval, Enabled: false, State: map[string]Item{}})
		e = a.save()
		a.mu.Unlock()
		if e != nil {
			http.Error(w, e.Error(), 500)
			return
		}
		redirect(w, r, "Job created paused. Enable when ready")
	case "toggle_job", "run_job", "delete_job":
		id := r.FormValue("id")
		a.mu.Lock()
		j := a.job(id)
		if j == nil {
			a.mu.Unlock()
			redirect(w, r, "Job not found")
			return
		}
		if action == "toggle_job" {
			j.Enabled = !j.Enabled
			e := a.save()
			a.mu.Unlock()
			if e != nil {
				http.Error(w, e.Error(), 500)
				return
			}
			redirect(w, r, "Job updated")
			return
		}
		if action == "delete_job" {
			if _, running := a.runs.Load(id); running {
				a.mu.Unlock()
				redirect(w, r, "Wait for sync to finish before deleting")
				return
			}
			for i := range a.cfg.Jobs {
				if a.cfg.Jobs[i].ID == id {
					a.cfg.Jobs = append(a.cfg.Jobs[:i], a.cfg.Jobs[i+1:]...)
					break
				}
			}
			e := a.save()
			a.mu.Unlock()
			if e != nil {
				http.Error(w, e.Error(), 500)
				return
			}
			redirect(w, r, "Mapping removed. Files left untouched")
			return
		}
		a.mu.Unlock()
		go func() {
			if e := a.syncJob(context.Background(), id); e != nil {
				log.Printf("job %s: %v", id, e)
			}
		}()
		redirect(w, r, "Sync started (check status)")
	case "remove_account":
		id := r.FormValue("id")
		a.mu.Lock()
		for _, j := range a.cfg.Jobs {
			if j.AccountID == id {
				a.mu.Unlock()
				redirect(w, r, "Remove mappings for this account first")
				return
			}
		}
		for i := range a.cfg.Accounts {
			if a.cfg.Accounts[i].ID == id {
				a.cfg.Accounts = append(a.cfg.Accounts[:i], a.cfg.Accounts[i+1:]...)
				break
			}
		}
		e := a.save()
		a.mu.Unlock()
		if e != nil {
			http.Error(w, e.Error(), 500)
			return
		}
		redirect(w, r, "Account removed from this installation")
	default:
		http.Error(w, "Unknown action", 400)
	}
}
func isLoopbackHost(h string) bool {
	if strings.EqualFold(h, "localhost") {
		return true
	}
	ip := net.ParseIP(h)
	return ip != nil && ip.IsLoopback()
}
func validDriveID(s string) bool {
	if len(s) < 8 || len(s) > 256 {
		return false
	}
	for _, c := range s {
		if c != '-' && c != '_' && (c < '0' || c > '9') && (c < 'A' || c > 'Z') && (c < 'a' || c > 'z') {
			return false
		}
	}
	return true
}
func (a *App) oauthStart(w http.ResponseWriter, r *http.Request) {
	if !a.authenticated(r) {
		http.Error(w, "Unauthorized", 401)
		return
	}
	a.mu.Lock()
	id := a.cfg.OAuthID
	base := a.cfg.BaseURL
	secret := a.cfg.OAuthSecret
	a.mu.Unlock()
	if id == "" || secret == "" || base == "" {
		redirect(w, r, "Configure Google OAuth first")
		return
	}
	state := b64rand()
	a.mu.Lock()
	if a.oauth == nil {
		a.oauth = map[string]time.Time{}
	}
	for k, t := range a.oauth {
		if time.Since(t) > 10*time.Minute {
			delete(a.oauth, k)
		}
	}
	a.oauth[state] = time.Now()
	a.mu.Unlock()
	q := url.Values{"client_id": {id}, "redirect_uri": {base + "/oauth/callback"}, "response_type": {"code"}, "scope": {scope + " https://www.googleapis.com/auth/userinfo.email"}, "access_type": {"offline"}, "prompt": {"consent select_account"}, "state": {state}}
	http.Redirect(w, r, "https://accounts.google.com/o/oauth2/v2/auth?"+q.Encode(), http.StatusFound)
}
func (a *App) oauthCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	a.mu.Lock()
	t, ok := a.oauth[state]
	delete(a.oauth, state)
	id := a.cfg.OAuthID
	secret := a.cfg.OAuthSecret
	base := a.cfg.BaseURL
	a.mu.Unlock()
	if !ok || time.Since(t) > 10*time.Minute || !a.authenticated(r) {
		http.Error(w, "Invalid or expired OAuth state", 403)
		return
	}
	if r.URL.Query().Get("error") != "" {
		redirect(w, r, "Google authorization cancelled or rejected")
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code", 400)
		return
	}
	v := url.Values{"code": {code}, "client_id": {id}, "client_secret": {secret}, "redirect_uri": {base + "/oauth/callback"}, "grant_type": {"authorization_code"}}
	req, e := http.NewRequestWithContext(r.Context(), "POST", "https://oauth2.googleapis.com/token", strings.NewReader(v.Encode()))
	if e != nil {
		http.Error(w, e.Error(), 500)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, e := a.client.Do(req)
	if e != nil {
		http.Error(w, e.Error(), 502)
		return
	}
	defer resp.Body.Close()
	var tok tokenResponse
	if e = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&tok); e != nil || resp.StatusCode != 200 || tok.RefreshToken == "" {
		redirect(w, r, "Google token exchange failed or refresh token missing")
		return
	}
	req, e = http.NewRequestWithContext(r.Context(), "GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if e != nil {
		http.Error(w, e.Error(), 500)
		return
	}
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	resp2, e := a.client.Do(req)
	if e != nil {
		http.Error(w, e.Error(), 502)
		return
	}
	defer resp2.Body.Close()
	var profile struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
	}
	if resp2.StatusCode != 200 || json.NewDecoder(io.LimitReader(resp2.Body, 1<<20)).Decode(&profile) != nil || profile.Sub == "" {
		redirect(w, r, "Could not identify Google account")
		return
	}
	a.mu.Lock()
	found := false
	for i := range a.cfg.Accounts {
		if a.cfg.Accounts[i].ID == profile.Sub {
			a.cfg.Accounts[i].RefreshToken = tok.RefreshToken
			a.cfg.Accounts[i].AccessToken = tok.AccessToken
			a.cfg.Accounts[i].Expires = time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)
			found = true
			break
		}
	}
	if !found {
		a.cfg.Accounts = append(a.cfg.Accounts, Account{ID: profile.Sub, Email: profile.Email, Label: profile.Email, RefreshToken: tok.RefreshToken, AccessToken: tok.AccessToken, Expires: time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)})
	}
	e = a.save()
	a.mu.Unlock()
	if e != nil {
		http.Error(w, e.Error(), 500)
		return
	}
	redirect(w, r, "Google account connected")
}

var ui = template.Must(template.New("app").Parse(newUI))

func (a *App) updateStatus(w http.ResponseWriter,r *http.Request){
 if r.Method!=http.MethodGet{http.Error(w,"Method not allowed",http.StatusMethodNotAllowed);return}
 if !a.authenticated(r){http.Error(w,"Unauthorized",http.StatusUnauthorized);return}
 ctx,cancel:=context.WithTimeout(r.Context(),15*time.Second);defer cancel()
 release,err:=latestRelease(ctx)
 w.Header().Set("Content-Type","application/json; charset=utf-8")
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("X-Content-Type-Options","nosniff")
 current:=strings.TrimPrefix(appVersion,"v")
 if err!=nil{_ = json.NewEncoder(w).Encode(map[string]any{"current":appVersion,"error":err.Error()});return}
 latest:=strings.TrimPrefix(release.Tag,"v")
 _ = json.NewEncoder(w).Encode(map[string]any{"current":appVersion,"latest":release.Tag,"available":latest!=current,"url":"https://github.com/gigabytegrove/syncria/releases/latest"})
}
func (a *App) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	a.mu.Lock()
	cfg := a.cfg
	cfg.Accounts = append([]Account(nil), a.cfg.Accounts...)
	cfg.Jobs = append([]Job(nil), a.cfg.Jobs...)
	cfg.Shares = append([]StorageShare(nil), a.cfg.Shares...)
	a.mu.Unlock()
	for i:=range cfg.Shares {cfg.Shares[i].Mounted=shareMounted(cfg.Shares[i].Local)}
	base := cfg.BaseURL
	if base == "" {
		base = "http://localhost:9764"
	}
	data := struct {
		Version, Message, ClientID, BaseURL, Callback string
		Configured, Logged                            bool
		Accounts                                      []Account
		Jobs                                          []Job
		Shares                                        []StorageShare
	}{appVersion, r.URL.Query().Get("msg"), cfg.OAuthID, base, base + "/oauth/callback", cfg.PasswordHash != "", a.authenticated(r), cfg.Accounts, cfg.Jobs, cfg.Shares}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; connect-src 'self'; style-src 'unsafe-inline'; form-action 'self'; base-uri 'none'; frame-ancestors 'none'; script-src 'unsafe-inline'")
	_ = ui.Execute(w, data)
}
func (a *App) scheduler(ctx context.Context) {
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
		a.mu.Lock()
		jobs := append([]Job(nil), a.cfg.Jobs...)
		a.mu.Unlock()
		for _, j := range jobs {
			if j.Enabled && (j.LastRun.IsZero() || time.Since(j.LastRun) >= time.Duration(j.Interval)*time.Minute) {
				id := j.ID
				go func() {
					if e := a.syncJob(ctx, id); e != nil && e.Error() != "job already running" {
						log.Printf("sync job %s: %v", id, e)
					}
				}()
			}
		}
	}
}
func restartSyncria(dataDir string){
    if os.Getenv("SYNCRIA_SUPERVISED")=="1" {os.Exit(75)}
    binary:=filepath.Join(dataDir,"runtime","current")
    cmd:=exec.Command("sh",append([]string{"-c","sleep 1; exec \"$@\"", "syncria",binary},os.Args[1:]...)...)
    cmd.Stdout=os.Stdout;cmd.Stderr=os.Stderr;cmd.Stdin=os.Stdin
    if err:=cmd.Start();err!=nil{log.Printf("update restart failed: %v",err);return}
    os.Exit(0)
}
func main() {
	listen := flag.String("listen", "127.0.0.1:9764", "Web UI listen address (use reverse proxy for external access)")
	dataDir := flag.String("data", "", "Configuration/data directory")
	flag.Parse()
	dir := *dataDir
	if dir == "" {
		if u, e := os.UserConfigDir(); e == nil {
			dir = filepath.Join(u, "gdsync")
		} else {
			log.Fatal(e)
		}
	}
	if e := os.MkdirAll(dir, 0700); e != nil {
		log.Fatal(e)
	}
	file := filepath.Join(dir, "state.json")
	a := &App{file: file, client: &http.Client{Timeout: 45 * time.Minute}, oauth: map[string]time.Time{}}
	if b, e := os.ReadFile(file); e == nil {
		if e = json.Unmarshal(b, &a.cfg); e != nil {
			log.Fatal(e)
		}
	} else if !os.IsNotExist(e) {
		log.Fatal(e)
	}
	if a.cfg.Secret == "" {
		a.cfg.Secret = b64rand()
		if e := a.save(); e != nil {
			log.Fatal(e)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go a.scheduler(ctx)
	go a.restoreShares()
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.home)
	mux.HandleFunc("/api/browse/drive", a.browseDrive)
	mux.HandleFunc("/api/update/status", a.updateStatus)
	mux.HandleFunc("/api/browse/local", a.browseLocal)
	mux.HandleFunc("/api/browse/local/create", a.createLocalFolder)
	mux.HandleFunc("/api/shares/action", a.shareAction)
	mux.HandleFunc("/action", a.action)
	mux.HandleFunc("/oauth/start", a.oauthStart)
	mux.HandleFunc("/oauth/callback", a.oauthCallback)
	server := &http.Server{Addr: *listen, Handler: mux, ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20}
	log.Printf("Syncria %s listening on %s (data: %s, OS: %s)", appVersion, *listen, dir, runtime.GOOS)
	if e := server.ListenAndServe(); e != nil && e != http.ErrServerClosed {
		log.Fatal(e)
	}
}