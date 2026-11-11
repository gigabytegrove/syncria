// Package syncpolicy enforces locally evaluated job-time and throughput limits.
// All time windows are interpreted in an explicit IANA timezone, including DST.
package syncpolicy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"time"
)

var ErrOutsideWindow = errors.New("synchronization outside scheduled window")

type Window struct {
	Timezone    string         `json:"timezone"`
	Days        []time.Weekday `json:"days"`
	StartMinute int            `json:"start_minute"`
	EndMinute   int            `json:"end_minute"`
}

type Policy struct {
	// KiBPerSecond=0 means unrestricted; >0 caps average transfer throughput.
	KiBPerSecond int      `json:"kib_per_second"`
	Windows      []Window `json:"windows,omitempty"`
}

func (p Policy) Validate() error {
	if p.KiBPerSecond < 0 || p.KiBPerSecond > 1024*1024 {
		return errors.New("rate outside allowed limits")
	}
	if len(p.Windows) > 32 {
		return errors.New("too many schedule windows")
	}
	for _, w := range p.Windows {
		if w.Timezone == "" {
			return errors.New("explicit timezone required")
		}
		if _, err := time.LoadLocation(w.Timezone); err != nil {
			return fmt.Errorf("unknown timezone %q: %w", w.Timezone, err)
		}
		if w.StartMinute < 0 || w.StartMinute >= 1440 || w.EndMinute < 0 || w.EndMinute >= 1440 || w.StartMinute == w.EndMinute {
			return errors.New("invalid schedule window")
		}
		if len(w.Days) == 0 || len(w.Days) > 7 {
			return errors.New("schedule requires weekdays")
		}
		seen := map[time.Weekday]bool{}
		for _, d := range w.Days {
			if d < time.Sunday || d > time.Saturday || seen[d] {
				return errors.New("invalid or duplicate weekday")
			}
			seen[d] = true
		}
	}
	return nil
}

// Allowed returns true when at least one configured window covers the instant.
// Cross-midnight schedules inherit the previous day's weekday through midnight.
func (p Policy) Allowed(instant time.Time) bool {
	if len(p.Windows) == 0 {
		return true
	}
	for _, w := range p.Windows {
		tz, err := time.LoadLocation(w.Timezone)
		if err != nil {
			continue
		}
		local := instant.In(tz)
		minute := local.Hour()*60 + local.Minute()
		currentDay, previousDay := local.Weekday(), (local.Weekday()+6)%7
		days := map[time.Weekday]bool{}
		for _, d := range w.Days {
			days[d] = true
		}
		if w.StartMinute < w.EndMinute {
			if days[currentDay] && minute >= w.StartMinute && minute < w.EndMinute {
				return true
			}
		} else if (days[currentDay] && minute >= w.StartMinute) || (days[previousDay] && minute < w.EndMinute) {
			return true
		}
	}
	return false
}

// RateReader uses a monotonic clock to throttle a stream in bounded chunks.
// It preserves Read/Close and cancellation semantics. No payload buffering.
type RateReader struct {
	Source         io.ReadCloser
	BytesPerSecond int64
	Context        context.Context
	Allowed        func(time.Time) bool
	next           time.Time
}

func (r *RateReader) Read(p []byte) (int, error) {
	if r.Source == nil {
		return 0, errors.New("missing source")
	}
	ctx := r.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if r.Allowed != nil && !r.Allowed(time.Now()) {
		return 0, ErrOutsideWindow
	}
	if len(p) == 0 {
		return r.Source.Read(p)
	}
	if r.BytesPerSecond <= 0 {
		return r.Source.Read(p)
	}
	// Bound each read to 64 KiB for prompt cancellation and limited burst.
	const maxChunk = 64 << 10
	if len(p) > maxChunk {
		p = p[:maxChunk]
	}
	// The rate is bytes/sec. A single chunk cannot exceed one second's budget.
	if r.BytesPerSecond < int64(len(p)) {
		p = p[:int(r.BytesPerSecond)]
	}
	if len(p) == 0 {
		return 0, nil
	}
	if !r.next.IsZero() {
		delay := time.Until(r.next)
		if delay > 0 {
			timer := time.NewTimer(delay)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			case <-timer.C:
			}
		}
	}
	n, err := r.Source.Read(p)
	if n > 0 {
		now := time.Now()
		if r.next.Before(now) {
			r.next = now
		}
		// nanoseconds / byte is rounded up, avoiding an unbounded speed boost.
		duration := time.Duration((int64(n)*int64(time.Second) + r.BytesPerSecond - 1) / r.BytesPerSecond)
		r.next = r.next.Add(duration)
	}
	return n, err
}
func (r *RateReader) Close() error {
	if r.Source == nil {
		return nil
	}
	return r.Source.Close()
}

func CanonicalDays(days []time.Weekday) []time.Weekday {
	v := append([]time.Weekday(nil), days...)
	sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })
	return v
}
