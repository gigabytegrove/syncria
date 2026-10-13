package main

import "os"

type DashboardTotals struct {
    Runs int64
    Failures int64
    Uploaded int64
    Downloaded int64
    UploadBytes int64
    DownloadBytes int64
    TrackedFiles int64
    Active int64
}

func summarizeMetrics(jobs []Job) DashboardTotals {
    var m DashboardTotals
    for _, j := range jobs {
        m.Runs += j.SuccessfulRuns+j.FailedRuns
        m.Failures += j.FailedRuns
        m.Uploaded += j.TotalUploads
        m.Downloaded += j.TotalDownloads
        m.UploadBytes += j.TotalUploadBytes
        m.DownloadBytes += j.TotalDownloadBytes
        if j.Enabled {m.Active++}
        for _, item := range j.State {if !item.Folder {m.TrackedFiles++}}
    }
    return m
}

func fileSize(path string) int64 {
    st, err := os.Stat(path)
    if err != nil || !st.Mode().IsRegular() {return 0}
    return st.Size()
}

func recordTransfer(j *Job, toDestination bool, size int64) {
    if size < 0 {size = 0}
    if toDestination {
        j.LastUploads++
        j.LastUploadBytes += size
        j.TotalUploads++
        j.TotalUploadBytes += size
    } else {
        j.LastDownloads++
        j.LastDownloadBytes += size
        j.TotalDownloads++
        j.TotalDownloadBytes += size
    }
}
