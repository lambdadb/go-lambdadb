package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUnixMilliTimeRoundTrip(t *testing.T) {
	wantMilliseconds := int64(1788336000123)
	original := UnixMilliTimeFromMilliseconds(wantMilliseconds)

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if got := string(data); got != "1788336000123" {
		t.Fatalf("Marshal() = %s, want 1788336000123", got)
	}

	var decoded UnixMilliTime
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got := decoded.UnixMilli(); got != wantMilliseconds {
		t.Fatalf("UnixMilli() = %d, want %d", got, wantMilliseconds)
	}
}

func TestUnixMilliTimeFrom(t *testing.T) {
	want := time.Date(2026, time.September, 2, 16, 0, 0, 123000000, time.UTC)
	got := UnixMilliTimeFrom(want)
	if !got.Equal(want) {
		t.Fatalf("UnixMilliTimeFrom() = %s, want %s", got.Time, want)
	}
}
