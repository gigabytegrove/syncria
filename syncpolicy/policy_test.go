package syncpolicy

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"
)

func TestWindowTimeZoneAndOvernight(t *testing.T) {
	p := Policy{Windows: []Window{{Timezone: "America/New_York", Days: []time.Weekday{time.Friday}, StartMinute: 22 * 60, EndMinute: 2 * 60}}}
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	check := func(day int, hour int, want bool) {
		t.Helper()
		at := time.Date(2026, 10, day, hour, 30, 0, 0, time.FixedZone("UTC", 0))
		if got := p.Allowed(at); got != want {
			t.Errorf("at %v allowed %v want %v", at, got, want)
		}
	}
	// Friday October 23 2026 22:30 EDT = Saturday Oct 24 02:30 UTC.
	check(24, 3, true)
	check(24, 5, true)
	check(24, 7, false)
	check(23, 20, false)
}
func TestDSTSpringForwardAndFallBack(t *testing.T) {
	p := Policy{Windows: []Window{{Timezone: "America/New_York", Days: []time.Weekday{time.Sunday}, StartMinute: 60, EndMinute: 180}}}
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, minute := range []int{6*60 + 15, 6*60 + 45} {
		if !p.Allowed(time.Date(2026, 3, 8, minute/60, minute%60, 0, 0, time.UTC)) {
			t.Fatal("spring forward window improperly closed")
		}
	}
	// 2026-11-01 01:30 local occurs in both EDT and EST, both are in the window.
	for _, h := range []int{5, 6} {
		if !p.Allowed(time.Date(2026, 11, 1, h, 30, 0, 0, time.UTC)) {
			t.Fatal("fall-back overlap not covered")
		}
	}
}
func TestInvalidSchedules(t *testing.T) {
	for _, p := range []Policy{{KiBPerSecond: -1}, {KiBPerSecond: 1048577}, {Windows: []Window{{Days: []time.Weekday{time.Monday}, Timezone: "Nowhere/Unknown", StartMinute: 10, EndMinute: 20}}}, {Windows: []Window{{Days: []time.Weekday{time.Monday}, Timezone: "UTC", StartMinute: 10, EndMinute: 10}}}, {Windows: []Window{{Days: []time.Weekday{time.Monday, time.Monday}, Timezone: "UTC", StartMinute: 10, EndMinute: 20}}}} {
		if err := p.Validate(); err == nil {
			t.Fatalf("invalid schedule accepted %+v", p)
		}
	}
	if !(Policy{}).Allowed(time.Now()) {
		t.Fatal("default policy must allow")
	}
}
func TestRateReaderBoundsAndClose(t *testing.T) {
	payload := bytes.Repeat([]byte("abcde"), 4096)
	r := &RateReader{Source: io.NopCloser(bytes.NewReader(payload)), BytesPerSecond: 128 * 1024, Context: context.Background()}
	t0 := time.Now()
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatal("rate reader changed payload")
	}
	if time.Since(t0) < 130*time.Millisecond {
		t.Fatal("rate enforcement ineffective")
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
}
func TestRateReaderCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	r := &RateReader{Source: io.NopCloser(bytes.NewReader(make([]byte, 4096))), BytesPerSecond: 128, Context: ctx}
	b := make([]byte, 4096)
	if _, err := r.Read(b); err != nil {
		t.Fatal(err)
	}
	cancel()
	if _, err := r.Read(b); !errors.Is(err, context.Canceled) {
		t.Fatal("cancel not enforced", err)
	}
}

func TestPolicyStopsTransferAfterWindowCloses(t *testing.T) {
	allowed := true
	r := &RateReader{Source: io.NopCloser(bytes.NewReader([]byte("abcdef"))), BytesPerSecond: 0, Context: context.Background(), Allowed: func(time.Time) bool { return allowed }}
	b := make([]byte, 3)
	n, err := r.Read(b)
	if err != nil || n != 3 {
		t.Fatalf("initial read %v n=%d", err, n)
	}
	allowed = false
	if _, err = r.Read(b); !errors.Is(err, ErrOutsideWindow) {
		t.Fatalf("window expired while reading: %v", err)
	}
}
