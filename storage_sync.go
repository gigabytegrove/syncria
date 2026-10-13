package main

import (
    "context"
    "crypto/md5"
    "encoding/hex"
    "errors"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "sort"
    "strings"
)

// checkedStoragePath resolves paths, enforces the configured storage boundary,
// and prevents writes to an accidentally unmounted managed network share.
func (a *App) checkedStoragePath(raw string) (string, error) {
    if raw == "" || !filepath.IsAbs(raw) {
        return "", errors.New("Choose an absolute storage directory")
    }
    root, err := resolvedPath(browseStorageRoot())
    if err != nil { return "", fmt.Errorf("Storage root unavailable: %w", err) }
    path, err := resolvedPath(filepath.Clean(raw))
    if err != nil { return "", fmt.Errorf("Storage folder unavailable: %w", err) }
    if !contained(root, path) { return "", errors.New("Directory is outside available storage") }
    st, err := os.Stat(path)
    if err != nil || !st.IsDir() { return "", errors.New("Storage directory must exist and be accessible") }
    if f, err := os.Open(path); err != nil { return "", fmt.Errorf("Storage directory cannot be opened: %w", err) } else { _ = f.Close() }
    a.mu.Lock()
    shares := append([]StorageShare(nil), a.cfg.Shares...)
    a.mu.Unlock()
    for _, share := range shares {
        if path == share.Local || strings.HasPrefix(path, share.Local+string(os.PathSeparator)) {
            if !shareMounted(share.Local) { return "", fmt.Errorf("Network share %s is disconnected; reconnect before syncing", share.Label) }
        }
    }
    return path, nil
}

func pathsOverlap(a, b string) bool {
    a = filepath.Clean(a); b = filepath.Clean(b)
    return a == b || strings.HasPrefix(a, b+string(os.PathSeparator)) || strings.HasPrefix(b, a+string(os.PathSeparator))
}

func copyStorageFile(src, dst, expectedMD5 string) (int64, error) {
    source, err := os.Open(src)
    if err != nil { return 0, err }
    defer source.Close()
    stat, err := source.Stat()
    if err != nil { return 0, err }
    if !stat.Mode().IsRegular() { return 0, errors.New("Source is not a regular file") }
    if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil { return 0, err }
    temp, err := os.CreateTemp(filepath.Dir(dst), ".syncria-transfer-*")
    if err != nil { return 0, err }
    defer os.Remove(temp.Name())
    digest := md5.New()
    n, err := io.Copy(io.MultiWriter(temp, digest), source)
    if err != nil { _ = temp.Close(); return 0, err }
    if expectedMD5 != "" && hex.EncodeToString(digest.Sum(nil)) != expectedMD5 {
        _ = temp.Close()
        return 0, errors.New("Source changed during copy; destination left untouched")
    }
    if err := temp.Sync(); err != nil { _ = temp.Close(); return 0, err }
    if err := temp.Close(); err != nil { return 0, err }
    if err := os.Rename(temp.Name(), dst); err != nil { return 0, err }
    return n, nil
}

// executeStorage syncs two already-mounted filesystem directories.
// First-run divergent files and post-sync deletions are blocked, not overwritten.
func (a *App) executeStorage(ctx context.Context, j *Job) error {
    left, err := a.checkedStoragePath(j.RemoteID)
    if err != nil { return err }
    right, err := a.checkedStoragePath(j.Local)
    if err != nil { return err }
    if pathsOverlap(left, right) { return errors.New("Source and destination must not overlap") }
    sources, err := localTree(left)
    if err != nil { return fmt.Errorf("Source scan: %w", err) }
    targets, err := localTree(right)
    if err != nil { return fmt.Errorf("Destination scan: %w", err) }
    union := make(map[string]bool, len(sources)+len(targets)+len(j.State))
    for k := range sources { union[k] = true }
    for k := range targets { union[k] = true }
    for k := range j.State { union[k] = true }
    keys := make([]string, 0, len(union))
    for k := range union { keys = append(keys, k) }
    sort.Slice(keys, func(i, k int) bool {
        di, dk := depthKey(keys[i]), depthKey(keys[k])
        if di != dk { return di < dk }
        return keys[i] < keys[k]
    })
    next := make(map[string]Item, len(union))
    for k,v := range j.State { next[k] = v }
    var uploaded, downloaded int64
    for i, rel := range keys {
        if err := ctx.Err(); err != nil { return err }
        if j.itemHook!=nil{j.itemHook(int64(i),int64(len(keys)),rel)}
        src, hasSource := sources[rel]
        dst, hasTarget := targets[rel]
        prior, tracked := j.State[rel]
        sourcePath, err := localPath(left, rel)
        if err != nil { return err }
        targetPath, err := localPath(right, rel)
        if err != nil { return err }
        if !hasSource && !hasTarget {
            delete(next, rel)
            if j.itemHook!=nil{j.itemHook(int64(i+1),int64(len(keys)),rel)}
            continue
        }
        if hasSource && hasTarget && src.Folder != dst.Folder {
            return fmt.Errorf("File/folder conflict: %s", rel)
        }
        if tracked && (!hasSource || !hasTarget) {
            return fmt.Errorf("Deletion detected for %s: review manually before resuming (no files removed)", rel)
        }
        isDir := (hasSource && src.Folder) || (hasTarget && dst.Folder)
        if isDir {
            if !hasSource { if err := os.MkdirAll(sourcePath,0700);err != nil{return err} }
            if !hasTarget { if err := os.MkdirAll(targetPath,0700);err != nil{return err} }
            next[rel] = Item{Folder:true}
            if j.itemHook!=nil{j.itemHook(int64(i+1),int64(len(keys)),rel)}
            continue
        }
        switch {
        case hasSource && hasTarget && src.LocalMD5 == dst.LocalMD5:
            next[rel] = Item{LocalMD5:src.LocalMD5,MD5:dst.LocalMD5}
        case hasSource && hasTarget:
            if !tracked { return fmt.Errorf("Initial file conflict at %s: neither side overwritten", rel) }
            sourceChanged := src.LocalMD5 != prior.LocalMD5
            targetChanged := dst.LocalMD5 != prior.MD5
            if sourceChanged == targetChanged { return fmt.Errorf("Conflicting changes at %s: neither side overwritten", rel) }
            if sourceChanged {
                n, err := copyStorageFile(sourcePath,targetPath,src.LocalMD5)
                if err != nil{return err}
                recordTransfer(j, true, n); uploaded++
                next[rel] = Item{LocalMD5:src.LocalMD5,MD5:src.LocalMD5}
            } else {
                n, err := copyStorageFile(targetPath,sourcePath,dst.LocalMD5)
                if err != nil{return err}
                recordTransfer(j, false, n); downloaded++
                next[rel] = Item{LocalMD5:dst.LocalMD5,MD5:dst.LocalMD5}
            }
        case hasSource:
            n, err := copyStorageFile(sourcePath,targetPath,src.LocalMD5)
            if err != nil{return err}
            recordTransfer(j,true,n); uploaded++
            next[rel] = Item{LocalMD5:src.LocalMD5,MD5:src.LocalMD5}
        case hasTarget:
            n, err := copyStorageFile(targetPath,sourcePath,dst.LocalMD5)
            if err != nil{return err}
            recordTransfer(j,false,n); downloaded++
            next[rel] = Item{LocalMD5:dst.LocalMD5,MD5:dst.LocalMD5}
        }
        j.State = next
        if j.itemHook!=nil{j.itemHook(int64(i+1),int64(len(keys)),rel)}
    }
    j.State = next
    if uploaded == 0 && downloaded == 0 {j.LastSummary="No file changes"} else {
        j.LastSummary=fmt.Sprintf("Source → destination: %d files, destination → source: %d files", uploaded, downloaded)
    }
    return nil
}
